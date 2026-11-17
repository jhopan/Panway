package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newBodyLimitTestRouter(maxBytes int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BodyLimit(maxBytes))
	r.POST("/v1/chat/completions", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusRequestEntityTooLarge, err.Error())
			return
		}
		c.String(http.StatusOK, string(body))
	})
	return r
}

func TestBodyLimitRejectsDeclaredOversizeBody(t *testing.T) {
	body := strings.Repeat("a", 64)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	w := httptest.NewRecorder()
	newBodyLimitTestRouter(16).ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("harus 413, dapat %d", w.Code)
	}
}

func TestBodyLimitRejectsChunkedOversizeBody(t *testing.T) {
	body := strings.Repeat("a", 64)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	// A chunked body has no Content-Length, so only MaxBytesReader can stop it.
	req.ContentLength = -1
	w := httptest.NewRecorder()
	newBodyLimitTestRouter(16).ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("harus 413, dapat %d", w.Code)
	}
}

func TestBodyLimitAllowsBodyUnderLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader("hello"))
	w := httptest.NewRecorder()
	newBodyLimitTestRouter(16).ServeHTTP(w, req)

	if w.Code != http.StatusOK || w.Body.String() != "hello" {
		t.Fatalf("body kecil harus lolos, dapat %d %q", w.Code, w.Body.String())
	}
}
