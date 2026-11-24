package ratio

import "github.com/jhopan/Panway/common/config"

// CacheHitRatio is the multiplier applied to prompt tokens the upstream
// served from its prompt cache. CACHE_HIT_DISCOUNT_PERCENT=80 means a cache
// hit is billed at 20% of the normal input price.
func CacheHitRatio() float64 {
	discount := config.CacheHitDiscountPercent
	if discount < 0 {
		discount = 0
	}
	if discount > 100 {
		discount = 100
	}
	return 1 - float64(discount)/100
}
