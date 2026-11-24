package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jhopan/Panway/common/env"

	"github.com/google/uuid"
)

var SystemName = "Panway"
var ServerAddress = "http://localhost:3000"
var Footer = ""
var Logo = ""
var TopUpLink = ""
var ChatLink = ""
var QuotaPerUnit = 500 * 1000.0 // $0.002 / 1K tokens
var DisplayInCurrencyEnabled = true
var DisplayTokenStatEnabled = true

// --- PanWallet integration ---
// PanWallet owns the customer's money. Panway only converts IDR into quota
// (and logs it). The quota unit is 1 IDR by default: QuotaPerUnit=1 means
// 1 quota unit == Rp1.
var PanWalletEnabled = env.Bool("PANWALLET_ENABLED", false)
var PanWalletURL = env.String("PANWALLET_URL", "")
var PanWalletKey = env.String("PANWALLET_KEY", "")

// Quota per 1 IDR credited. Default 1 quota = Rp1.
var PanWalletQuotaPerIDR = env.Float64("PANWALLET_QUOTA_PER_IDR", 1)

// Public base URL of PanWallet for the customer to see their balance.
var PanWalletPublicURL = env.String("PANWALLET_PUBLIC_URL", "")

// --- Telegram verification ---
// Telegram is used to verify that a customer controls a chat, not as a
// passwordless login. The customer registers with email or Google, then
// confirms a code sent to their Telegram chat.
var TelegramVerificationEnabled = env.Bool("TELEGRAM_VERIFICATION_ENABLED", false)
var TelegramBotToken = env.String("TELEGRAM_BOT_TOKEN", "")
var TelegramBotUsername = env.String("TELEGRAM_BOT_USERNAME", "")
var TelegramApiBaseURL = env.String("TELEGRAM_API_BASE_URL", "")

// Any options with "Secret", "Token" in its key won't be return by GetOptions

var SessionSecret = uuid.New().String()

// --- Production safety ---
//
// These settings exist so the gateway can be sold and run without relying on
// framework defaults that are convenient in development and unsafe in
// production.

// SessionCookieSecure marks the session cookie Secure, so browsers only send
// it over HTTPS. Turn it on once the console is served through TLS.
var SessionCookieSecure = env.Bool("SESSION_COOKIE_SECURE", false)

// TrustedProxies lists the reverse proxies whose X-Forwarded-For header is
// believed. Empty means no proxy is trusted and the client IP is taken from
// the TCP connection, which stops a caller from choosing its own rate limit
// bucket by spoofing the header. Set it to the proxy address (for example
// "127.0.0.1" for a local nginx) when Panway runs behind one.
var TrustedProxies = SplitList(env.String("TRUSTED_PROXIES", ""))

// CorsAllowedOrigins lists the browser origins allowed to call /v1 and
// /dashboard with credentials. Empty means same-origin only: no
// Access-Control-Allow-Origin header is sent. The upstream default allowed
// every origin together with credentials, which lets any website read a
// logged-in administrator's console data.
var CorsAllowedOrigins = SplitList(env.String("CORS_ALLOWED_ORIGINS", ""))

// MaxRequestBodyBytes caps one request body, including a gzip-decoded one.
// 0 disables the cap.
var MaxRequestBodyBytes = int64(env.Int("MAX_REQUEST_BODY_MB", 32)) * 1024 * 1024

// PricingCacheSeconds is how long the public /api/pricing answer is cached in
// process. Editing a price invalidates it immediately; the TTL only bounds
// how long a channel/ratio change made elsewhere stays invisible.
var PricingCacheSeconds = env.Int("PRICING_CACHE_SECONDS", 30)

// EstimatedUsageMode decides how usage the upstream flagged as estimated is
// billed. "local" (default) rebuilds the numbers from Panway's own tokenizers
// plus the hidden prompt overhead learned from exact usage; "upstream" bills
// the upstream estimate as it arrived, which matches the upstream meter but
// can be several times the real prompt size.
var EstimatedUsageMode = env.String("ESTIMATED_USAGE_MODE", "local")

// --- SQLite operations ---
// SQLITE_BACKUP_ENABLED turns on the scheduled VACUUM INTO snapshot. Keep it
// off only when an external backup tool already copies the database.
var SQLiteBackupEnabled = env.Bool("SQLITE_BACKUP_ENABLED", false)
var SQLiteBackupDir = env.String("SQLITE_BACKUP_DIR", "data/backups")
var SQLiteBackupIntervalHours = env.Int("SQLITE_BACKUP_INTERVAL_HOURS", 24)
var SQLiteBackupKeep = env.Int("SQLITE_BACKUP_KEEP", 7)

// LogRetentionDays deletes operation logs older than this many days. 0 keeps
// logs forever.
var LogRetentionDays = env.Int("LOG_RETENTION_DAYS", 0)

// LowBalanceThreshold enables a Telegram reminder when a verified customer's
// quota drops below this number of quota units. 0 disables it.
var LowBalanceThreshold = int64(env.Int("LOW_BALANCE_THRESHOLD", 0))
var LowBalanceNotifyIntervalHours = env.Int("LOW_BALANCE_NOTIFY_INTERVAL_HOURS", 24)

// CacheHitDiscountPercent is the discount applied to prompt tokens the
// upstream served from its prompt cache. 80 means a cache hit costs 20% of
// the normal input price. 0 bills cache hits like normal input tokens.
var CacheHitDiscountPercent = env.Int("CACHE_HIT_DISCOUNT_PERCENT", 80)
var MetricsEnabled = env.Bool("ENABLE_METRICS", false)
var MetricsToken = env.String("METRICS_TOKEN", "")

// Token keys are stored encrypted, with a keyed hash for lookup. The key comes
// from TOKEN_ENCRYPTION_KEY when set; otherwise a 32-byte key file is created
// once at TOKEN_KEY_FILE (default data/token.key). Losing that value makes
// stored keys unreadable, so back it up with the database.
var TokenEncryptionKey = env.String("TOKEN_ENCRYPTION_KEY", "")
var TokenKeyFile = env.String("TOKEN_KEY_FILE", "data/token.key")

// HstsMaxAge enables Strict-Transport-Security when greater than zero. Keep it
// disabled until the console is permanently served over HTTPS.
var HstsMaxAge = env.Int("HSTS_MAX_AGE", 0)
var HstsIncludeSubdomains = env.Bool("HSTS_INCLUDE_SUBDOMAINS", false)

// SplitList splits a comma-separated env value, dropping empty entries.
func SplitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}

var OptionMap map[string]string
var OptionMapRWMutex sync.RWMutex

var ItemsPerPage = 10
var MaxRecentItems = 100

var PasswordLoginEnabled = true
var PasswordRegisterEnabled = true
var EmailVerificationEnabled = false
var GitHubOAuthEnabled = false
var OidcEnabled = false
var WeChatAuthEnabled = false
var TurnstileCheckEnabled = false
var RegisterEnabled = true

var EmailDomainRestrictionEnabled = false
var EmailDomainWhitelist = []string{
	"gmail.com",
	"163.com",
	"126.com",
	"qq.com",
	"outlook.com",
	"hotmail.com",
	"icloud.com",
	"yahoo.com",
	"foxmail.com",
}

var DebugEnabled = strings.ToLower(os.Getenv("DEBUG")) == "true"
var DebugSQLEnabled = strings.ToLower(os.Getenv("DEBUG_SQL")) == "true"
var MemoryCacheEnabled = strings.ToLower(os.Getenv("MEMORY_CACHE_ENABLED")) == "true"

var LogConsumeEnabled = true

var SMTPServer = ""
var SMTPPort = 587
var SMTPAccount = ""
var SMTPFrom = ""
var SMTPToken = ""

var GitHubClientId = ""
var GitHubClientSecret = ""

var LarkClientId = ""
var LarkClientSecret = ""

var OidcClientId = ""
var OidcClientSecret = ""
var OidcWellKnown = ""
var OidcAuthorizationEndpoint = ""
var OidcTokenEndpoint = ""
var OidcUserinfoEndpoint = ""

var WeChatServerAddress = ""
var WeChatServerToken = ""
var WeChatAccountQRCodeImageURL = ""

var MessagePusherAddress = ""
var MessagePusherToken = ""

var TurnstileSiteKey = ""
var TurnstileSecretKey = ""

var QuotaForNewUser int64 = 0
var QuotaForInviter int64 = 0
var QuotaForInvitee int64 = 0
var ChannelDisableThreshold = 5.0
var AutomaticDisableChannelEnabled = false
var AutomaticEnableChannelEnabled = false
var QuotaRemindThreshold int64 = 1000
var PreConsumedQuota int64 = 500
var ApproximateTokenEnabled = false
var RetryTimes = env.Int("RETRY_TIMES", 0)

var RootUserEmail = ""

var IsMasterNode = os.Getenv("NODE_TYPE") != "slave"

var requestInterval, _ = strconv.Atoi(os.Getenv("POLLING_INTERVAL"))
var RequestInterval = time.Duration(requestInterval) * time.Second

var SyncFrequency = env.Int("SYNC_FREQUENCY", 10*60) // unit is second

var BatchUpdateEnabled = false
var BatchUpdateInterval = env.Int("BATCH_UPDATE_INTERVAL", 5)

var RelayTimeout = env.Int("RELAY_TIMEOUT", 0) // unit is second

var GeminiSafetySetting = env.String("GEMINI_SAFETY_SETTING", "BLOCK_NONE")

var Theme = env.String("THEME", "default")
var ValidThemes = map[string]bool{
	"default": true,
}

// All duration's unit is seconds
// Shouldn't larger then RateLimitKeyExpirationDuration
var (
	GlobalApiRateLimitNum            = env.Int("GLOBAL_API_RATE_LIMIT", 480)
	GlobalApiRateLimitDuration int64 = 3 * 60

	GlobalWebRateLimitNum            = env.Int("GLOBAL_WEB_RATE_LIMIT", 240)
	GlobalWebRateLimitDuration int64 = 3 * 60

	UploadRateLimitNum            = 10
	UploadRateLimitDuration int64 = 60

	DownloadRateLimitNum            = 10
	DownloadRateLimitDuration int64 = 60

	CriticalRateLimitNum            = 20
	CriticalRateLimitDuration int64 = 20 * 60
)

var RateLimitKeyExpirationDuration = 20 * time.Minute

var EnableMetric = env.Bool("ENABLE_METRIC", false)
var MetricQueueSize = env.Int("METRIC_QUEUE_SIZE", 10)
var MetricSuccessRateThreshold = env.Float64("METRIC_SUCCESS_RATE_THRESHOLD", 0.8)
var MetricSuccessChanSize = env.Int("METRIC_SUCCESS_CHAN_SIZE", 1024)
var MetricFailChanSize = env.Int("METRIC_FAIL_CHAN_SIZE", 128)

var InitialRootToken = os.Getenv("INITIAL_ROOT_TOKEN")

var InitialRootAccessToken = os.Getenv("INITIAL_ROOT_ACCESS_TOKEN")

var GeminiVersion = env.String("GEMINI_VERSION", "v1")

var OnlyOneLogFile = env.Bool("ONLY_ONE_LOG_FILE", false)

var RelayProxy = env.String("RELAY_PROXY", "")
var UserContentRequestProxy = env.String("USER_CONTENT_REQUEST_PROXY", "")
var UserContentRequestTimeout = env.Int("USER_CONTENT_REQUEST_TIMEOUT", 30)

var EnforceIncludeUsage = env.Bool("ENFORCE_INCLUDE_USAGE", false)
var TestPrompt = env.String("TEST_PROMPT", "Output only your specific model name with no additional text.")
