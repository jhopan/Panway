package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common/config"
)

func newHeadersTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.Use(NoStoreAPI())
	handler := func(c *gin.Context) { c.String(http.StatusOK, "ok") }
	r.GET("/", handler)
	r.GET("/api/user/self", handler)
	r.GET("/v1/models", handler)
	r.GET("/dashboard/billing/usage", handler)
	return r
}

func TestSecurityHeadersOnEveryResponse(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	newHeadersTestRouter().ServeHTTP(w, req)

	expected := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}
	for header, want := range expected {
		if got := w.Header().Get(header); got != want {
			t.Fatalf("%s harus %q, dapat %q", header, want, got)
		}
	}
}

func TestHSTSOnlyWhenConfigured(t *testing.T) {
	previous := config.HstsMaxAge
	defer func() { config.HstsMaxAge = previous }()

	config.HstsMaxAge = 0
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	w := httptest.NewRecorder()
	newHeadersTestRouter().ServeHTTP(w, req)
	if got := w.Header().Get("Strict-Transport-Security"); got != "" {
		t.Fatalf("HSTS harus mati saat tidak dikonfigurasi, dapat %q", got)
	}

	config.HstsMaxAge = 31536000
	req = httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	w = httptest.NewRecorder()
	newHeadersTestRouter().ServeHTTP(w, req)
	if got := w.Header().Get("Strict-Transport-Security"); got != "max-age=31536000" {
		t.Fatalf("HSTS harus aktif di HTTPS, dapat %q", got)
	}
}

func TestAPIRoutesAreNotStored(t *testing.T) {
	for _, path := range []string{"/api/user/self", "/v1/models", "/dashboard/billing/usage"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		newHeadersTestRouter().ServeHTTP(w, req)

		if got := w.Header().Get("Cache-Control"); got != "no-store" {
			t.Fatalf("%s harus no-store, dapat %q", path, got)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	newHeadersTestRouter().ServeHTTP(w, req)
	if got := w.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("halaman non-API tidak perlu no-store, dapat %q", got)
	}
}
