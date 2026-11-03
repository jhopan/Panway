package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common/i18n"
)

func Language() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			lang = "en"
		}
		lower := strings.ToLower(lang)
		switch {
		case strings.HasPrefix(lower, "id"):
			lang = "id"
		case strings.HasPrefix(lower, "en"):
			lang = "en"
		default:
			lang = "en"
		}
		c.Set(i18n.ContextKey, lang)
		c.Next()
	}
}
