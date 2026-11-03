package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common"
	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/common/ctxkey"
	"github.com/jhopan/Panway/model"
)

// TokenRateLimit enforces the per-token rate limits: requests per minute,
// tokens per minute, and concurrent requests.
//
// It runs after TokenAuth, so the token id is already in the context. The
// request body is read to estimate the prompt tokens for the TPM check; the
// body is restored afterwards so the relay handler can still read it.
func TokenRateLimit() func(c *gin.Context) {
	return func(c *gin.Context) {
		tokenId := c.GetInt(ctxkey.TokenId)
		if tokenId == 0 {
			c.Next()
			return
		}
		token, err := model.GetTokenById(tokenId)
		if err != nil {
			// No token record (cache miss) — do not block the request.
			c.Next()
			return
		}

		limiter := common.GetTokenRateLimiter()

		rpm := token.RateLimitRPM
		if rpm <= 0 {
			rpm = config.GlobalApiRateLimitNum
		}
		if !limiter.AllowRequest(tokenId, rpm) {
			abortWithMessage(c, http.StatusTooManyRequests,
				"rate limit exceeded: token reached "+strconv.Itoa(rpm)+" requests per minute")
			return
		}

		release, ok := limiter.Acquire(tokenId, token.ConcurrencyLimit)
		if !ok {
			abortWithMessage(c, http.StatusTooManyRequests,
				"rate limit exceeded: too many concurrent requests for this token")
			return
		}
		defer release()

		c.Next()
	}
}
