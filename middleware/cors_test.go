package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common/config"
)

func newCorsTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS())
	handler := func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"success": true}) }
	r.GET("/v1/models", handler)
	r.POST("/v1/chat/completions", handler)
	return r
}

func TestCORSRejectsUnconfiguredOrigin(t *testing.T) {
	previous := config.CorsAllowedOrigins
	defer func() { config.CorsAllowedOrigins = previous }()
	config.CorsAllowedOrigins = []string{"https://console.example.com"}

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	w := httptest.NewRecorder()
	newCorsTestRouter().ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("origin tidak boleh diizinkan, dapat %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("credentials tidak boleh diizinkan, dapat %q", got)
	}
}

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	previous := config.CorsAllowedOrigins
	defer func() { config.CorsAllowedOrigins = previous }()
	config.CorsAllowedOrigins = []string{"https://console.example.com"}

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set("Origin", "https://console.example.com")
	w := httptest.NewRecorder()
	newCorsTestRouter().ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://console.example.com" {
		t.Fatalf("origin harus diizinkan, dapat %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("credentials harus diizinkan, dapat %q", got)
	}
}

func TestCORSPreflightEchoesRequestedHeaders(t *testing.T) {
	previous := config.CorsAllowedOrigins
	defer func() { config.CorsAllowedOrigins = previous }()
	config.CorsAllowedOrigins = []string{"https://console.example.com"}

	requested := "authorization,content-type,x-stainless-os"
	req := httptest.NewRequest(http.MethodOptions, "/v1/chat/completions", nil)
	req.Header.Set("Origin", "https://console.example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", requested)
	w := httptest.NewRecorder()
	newCorsTestRouter().ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("preflight harus 204, dapat %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Headers"); got != requested {
		t.Fatalf("header yang diminta harus dipantulkan, dapat %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("Access-Control-Allow-Methods kosong")
	}
}

func TestCORSNoOriginsConfiguredSendsNoHeaders(t *testing.T) {
	previous := config.CorsAllowedOrigins
	defer func() { config.CorsAllowedOrigins = previous }()
	config.CorsAllowedOrigins = nil

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set("Origin", "https://console.example.com")
	w := httptest.NewRecorder()
	newCorsTestRouter().ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("tanpa konfigurasi tidak boleh ada header CORS, dapat %q", got)
	}
}
