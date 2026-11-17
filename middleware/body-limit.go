package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodyLimit caps the size of one request body.
//
// Without it a single caller can post a multi-gigabyte JSON body and make the
// relay allocate it in memory before the token is even checked. The cap is
// applied to the raw body; GzipDecodeMiddleware also caps the decompressed
// size, because a small gzip body can expand without bound.
//
// A body that grew past the limit is reported as 413 when the handler reads
// it, and immediately when Content-Length already exceeds the limit.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes <= 0 {
			c.Next()
			return
		}
		if c.Request.ContentLength > maxBytes {
			abortWithMessage(c, http.StatusRequestEntityTooLarge,
				fmt.Sprintf("request body is too large: the limit is %d bytes", maxBytes))
			return
		}
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
