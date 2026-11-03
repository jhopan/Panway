package controller

import (
	"github.com/jhopan/Panway/model"
)

// tokenTokenRateLimit returns the per-token TPM limit, or 0 when unlimited.
//
// The token record is read from the database on every request. That is a
// deliberate trade-off: the limit must take effect immediately after an admin
// changes it, and the token lookup is a single indexed row read.
func tokenTokenRateLimit(tokenId int) int {
	if tokenId == 0 {
		return 0
	}
	token, err := model.GetTokenById(tokenId)
	if err != nil || token == nil {
		return 0
	}
	return token.RateLimitTPM
}
