package middleware

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common/config"
)

// GzipDecodeMiddleware decompresses a gzip request body.
//
// The decompressed stream is capped like any other request body: the raw gzip
// body is already bounded by BodyLimit, but a few kilobytes of gzip can expand
// into gigabytes, so the size that reaches the handler has to be bounded too.
func GzipDecodeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			defer gzipReader.Close()

			// Replace the request body with the decompressed data, bounded to
			// the same size the compressed body is allowed to have.
			var body io.Reader = gzipReader
			if config.MaxRequestBodyBytes > 0 {
				body = http.MaxBytesReader(c.Writer, io.NopCloser(gzipReader), config.MaxRequestBodyBytes)
			}
			c.Request.Body = io.NopCloser(body)
		}

		// Continue processing the request
		c.Next()
	}
}
