package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common/config"
)

// CORS answers browser cross-origin requests only for the origins listed in
// CORS_ALLOWED_ORIGINS.
//
// The upstream default allowed every origin together with credentials, which
// lets any website send cookie-authenticated requests to the console and read
// the administrator's data. With no origin configured this middleware sends
// no CORS headers at all: the console, the landing page and the API are all
// served from the same origin, so nothing legitimate needs them.
//
// Browser SDK traffic (/v1) is the only place that normally needs an entry
// here, and only when the caller runs in a browser rather than on a server.
func CORS() gin.HandlerFunc {
	allowed := make(map[string]bool, len(config.CorsAllowedOrigins))
	for _, origin := range config.CorsAllowedOrigins {
		allowed[strings.TrimSuffix(origin, "/")] = true
	}
	return func(c *gin.Context) {
		origin := strings.TrimSuffix(c.GetHeader("Origin"), "/")
		if origin == "" || !allowed[origin] {
			c.Next()
			return
		}

		header := c.Writer.Header()
		header.Set("Access-Control-Allow-Origin", origin)
		header.Set("Access-Control-Allow-Credentials", "true")
		header.Add("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			// Echo the requested headers: browser SDKs send a long list of
			// x-stainless-* and openai-* headers that changes between
			// releases, so a fixed allowlist would break them later.
			if requested := c.GetHeader("Access-Control-Request-Headers"); requested != "" {
				header.Set("Access-Control-Allow-Headers", requested)
			}
			header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			header.Set("Access-Control-Max-Age", "43200")
			header.Add("Vary", "Access-Control-Request-Headers")
			header.Add("Vary", "Access-Control-Request-Method")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
