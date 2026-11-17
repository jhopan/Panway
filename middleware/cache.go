package middleware

import (
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// staticAssetExtensions lists the files whose name carries a content hash, so
// a one-week cache cannot go stale after a deploy.
var staticAssetExtensions = map[string]bool{
	".js":    true,
	".css":   true,
	".map":   true,
	".png":   true,
	".jpg":   true,
	".jpeg":  true,
	".gif":   true,
	".svg":   true,
	".ico":   true,
	".webp":  true,
	".avif":  true,
	".woff":  true,
	".woff2": true,
	".ttf":   true,
	".eot":   true,
}

// Cache sets Cache-Control for the frontend bundle.
//
// The previous version cached every path except "/" for a week, including
// SPA routes such as /console/token, which serve index.html. A browser that
// cached index.html for a week kept running the old app after a deploy. Only
// hashed assets get the long cache now; everything else is revalidated.
func Cache() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if requestPath != "/" && staticAssetExtensions[strings.ToLower(path.Ext(requestPath))] {
			c.Header("Cache-Control", "public, max-age=604800, immutable")
		} else {
			c.Header("Cache-Control", "no-cache")
		}
		c.Next()
	}
}
