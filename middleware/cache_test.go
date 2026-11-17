package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newCacheTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Cache())
	handler := func(c *gin.Context) { c.String(http.StatusOK, "ok") }
	r.GET("/", handler)
	r.GET("/console/token", handler)
	r.GET("/static/js/main.9f8e7d.js", handler)
	r.GET("/logo.png", handler)
	r.GET("/manifest.json", handler)
	return r
}

func TestCacheHashedAssetIsCachedForAWeek(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/static/js/main.9f8e7d.js", nil)
	w := httptest.NewRecorder()
	newCacheTestRouter().ServeHTTP(w, req)

	if got := w.Header().Get("Cache-Control"); got != "public, max-age=604800, immutable" {
		t.Fatalf("aset ber-hash harus cached lama, dapat %q", got)
	}
}

func TestCacheSPARouteIsRevalidated(t *testing.T) {
	for _, path := range []string{"/", "/console/token", "/manifest.json"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		newCacheTestRouter().ServeHTTP(w, req)

		if got := w.Header().Get("Cache-Control"); got != "no-cache" {
			t.Fatalf("%s harus no-cache, dapat %q", path, got)
		}
	}
}
