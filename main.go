package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"

	"github.com/jhopan/Panway/cmd/connect"
	"github.com/jhopan/Panway/common"
	"github.com/jhopan/Panway/common/client"
	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/common/i18n"
	"github.com/jhopan/Panway/common/logger"
	"github.com/jhopan/Panway/controller"
	"github.com/jhopan/Panway/middleware"
	"github.com/jhopan/Panway/model"
	"github.com/jhopan/Panway/relay/adaptor/openai"
	"github.com/jhopan/Panway/router"
)

//go:embed web/build/*
var buildFS embed.FS

func main() {
	if len(os.Args) > 1 && os.Args[1] == "connect" {
		os.Exit(connect.Run(os.Args[2:], os.Stdout))
	}
	common.Init()
	logger.SetupLogger()
	logger.SysLogf("Panway %s started", common.Version)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if os.Getenv("GIN_MODE") != gin.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	if config.DebugEnabled {
		logger.SysLog("running in debug mode")
	}

	// Initialize SQL Database
	model.InitDB()
	model.InitLogDB()
	model.MigrateTokenKeys()

	var err error
	err = model.CreateRootAccountIfNeed()
	if err != nil {
		logger.FatalLog("database init error: " + err.Error())
	}
	model.WarnIfDefaultRootPassword()
	defer func() {
		err := model.CloseDB()
		if err != nil {
			logger.FatalLog("failed to close database: " + err.Error())
		}
	}()

	// Initialize Redis
	err = common.InitRedisClient()
	if err != nil {
		logger.FatalLog("failed to initialize Redis: " + err.Error())
	}

	// Initialize options
	model.InitOptionMap()
	model.PrunePriceHistory(90)
	if config.PanWalletEnabled && (config.PanWalletURL == "" || config.PanWalletKey == "") {
		logger.SysError("PANWALLET_ENABLED is true but PANWALLET_URL or PANWALLET_KEY is empty; wallet top-up will fail until both are set")
	}
	model.StartSQLiteBackup(ctx)
	model.StartLogCleanup(ctx)
	model.StartLowBalanceNotifier(ctx)
	model.StartPlanExpiry(ctx)
	logger.SysLog(fmt.Sprintf("using theme %s", config.Theme))
	if common.RedisEnabled {
		// for compatibility with old versions
		config.MemoryCacheEnabled = true
	}
	if config.MemoryCacheEnabled {
		logger.SysLog("memory cache enabled")
		logger.SysLog(fmt.Sprintf("sync frequency: %d seconds", config.SyncFrequency))
		model.InitChannelCache()
	}
	if config.MemoryCacheEnabled {
		go model.SyncOptions(config.SyncFrequency)
		go model.SyncChannelCache(config.SyncFrequency)
	}
	if os.Getenv("CHANNEL_TEST_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("CHANNEL_TEST_FREQUENCY"))
		if err != nil {
			logger.FatalLog("failed to parse CHANNEL_TEST_FREQUENCY: " + err.Error())
		}
		go controller.AutomaticallyTestChannels(frequency)
	}
	if os.Getenv("BATCH_UPDATE_ENABLED") == "true" {
		config.BatchUpdateEnabled = true
		logger.SysLog("batch update enabled with interval " + strconv.Itoa(config.BatchUpdateInterval) + "s")
		model.InitBatchUpdater()
	}
	if config.EnableMetric {
		logger.SysLog("metric enabled, will disable channel if too much request failed")
	}
	openai.InitTokenEncoders()
	client.Init()

	// Initialize i18n
	if err := i18n.Init(); err != nil {
		logger.FatalLog("failed to initialize i18n: " + err.Error())
	}

	// Initialize HTTP server
	server := gin.New()
	server.Use(gin.Recovery())
	// This will cause SSE not to work!!!
	//server.Use(gzip.Gzip(gzip.DefaultCompression))
	server.Use(middleware.RequestId())
	server.Use(middleware.SecurityHeaders())
	server.Use(middleware.NoStoreAPI())
	server.Use(middleware.BodyLimit(config.MaxRequestBodyBytes))
	server.Use(middleware.Language())
	middleware.SetUpLogger(server)

	// Only trust X-Forwarded-For from the proxies named in TRUSTED_PROXIES.
	// Gin trusts every proxy by default, which lets a caller spoof its client
	// IP and walk around the global rate limits and the token subnet checks.
	if len(config.TrustedProxies) == 0 {
		if err := server.SetTrustedProxies(nil); err != nil {
			logger.FatalLog("failed to disable trusted proxies: " + err.Error())
		}
		logger.SysLog("trusted proxies: none, client IP is read from the connection")
	} else {
		if err := server.SetTrustedProxies(config.TrustedProxies); err != nil {
			logger.FatalLog("failed to set TRUSTED_PROXIES: " + err.Error())
		}
		logger.SysLog("trusted proxies: " + strings.Join(config.TrustedProxies, ", "))
	}

	// Initialize session store
	store := cookie.NewStore([]byte(config.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   config.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	server.Use(sessions.Sessions("session", store))

	router.SetRouter(server, buildFS)
	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}

	// ReadHeaderTimeout stops a slowloris client from holding a connection
	// open with an unfinished header. There is deliberately no WriteTimeout:
	// streaming completions legitimately stay open for minutes.
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           server,
		ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	logger.SysLogf("server started on http://localhost:%s", port)
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			logger.FatalLog("failed to start HTTP server: " + err.Error())
		}
	case <-ctx.Done():
		logger.SysLog("shutdown signal received, draining connections")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.SysError("graceful shutdown failed: " + err.Error())
		}
	}
}
