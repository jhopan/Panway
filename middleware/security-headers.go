package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common/config"
)

// apiPathPrefixes are the paths whose responses must never be stored by a
// browser or a shared proxy: they carry tokens, logs, balances, or per-user
// model lists.
var apiPathPrefixes = []string{"/api", "/v1", "/dashboard"}

func isAPIPath(requestPath string) bool {
	for _, prefix := range apiPathPrefixes {
		if requestPath == prefix || strings.HasPrefix(requestPath, prefix+"/") {
			return true
		}
	}
	return false
}

// NoStoreAPI marks API responses as not cacheable.
//
// The relay answers streaming responses as well, so this only sets the header
// and leaves the body untouched.
func NoStoreAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isAPIPath(c.Request.URL.Path) {
			c.Header("Cache-Control", "no-store")
		}
		c.Next()
	}
}

// SecurityHeaders sets the browser safety headers for every response.
//
// HSTS is opt-in through HSTS_MAX_AGE because Panway may be served over plain
// HTTP during evaluation, and a long HSTS lifetime on the wrong host is hard
// to undo. Set it only on the final HTTPS deployment.
func SecurityHeaders() gin.HandlerFunc {
	hsts := ""
	if config.HstsMaxAge > 0 {
		hsts = "max-age=" + strconv.Itoa(config.HstsMaxAge)
		if config.HstsIncludeSubdomains {
			hsts += "; includeSubDomains"
		}
	}
	return func(c *gin.Context) {
		header := c.Writer.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		header.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		if hsts != "" && (c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https") {
			header.Set("Strict-Transport-Security", hsts)
		}
		c.Next()
	}
}
