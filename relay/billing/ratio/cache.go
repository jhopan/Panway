package ratio

import (
	"encoding/json"
	"sync"

	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/common/logger"
)

// CacheHitRatio maps a model id to the fraction of the input price charged
// for prompt tokens served from the upstream prompt cache. 0.2 means a cache
// hit costs 20% of the normal input price. Models without an entry fall back
// to CACHE_HIT_DISCOUNT_PERCENT.
var (
	cacheHitRatioLock sync.RWMutex
	CacheHitRatio     = make(map[string]float64)
)

// DefaultCacheHitRatio is the fallback from CACHE_HIT_DISCOUNT_PERCENT.
func DefaultCacheHitRatio() float64 {
	discount := config.CacheHitDiscountPercent
	if discount < 0 {
		discount = 0
	}
	if discount > 100 {
		discount = 100
	}
	return 1 - float64(discount)/100
}

// GetCacheHitRatio returns the per-model cache fraction, or the global
// default when the model has no override.
func GetCacheHitRatio(model string) float64 {
	cacheHitRatioLock.RLock()
	value, ok := CacheHitRatio[model]
	cacheHitRatioLock.RUnlock()
	if !ok {
		return DefaultCacheHitRatio()
	}
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

// SetCacheHitRatio stores an override and returns the full map as JSON.
func SetCacheHitRatio(model string, value float64) (string, error) {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	cacheHitRatioLock.Lock()
	CacheHitRatio[model] = value
	jsonBytes, err := json.Marshal(CacheHitRatio)
	cacheHitRatioLock.Unlock()
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// DeleteCacheHitRatio removes an override so the model uses the global
// discount again. It returns the map as JSON.
func DeleteCacheHitRatio(model string) (string, error) {
	cacheHitRatioLock.Lock()
	delete(CacheHitRatio, model)
	jsonBytes, err := json.Marshal(CacheHitRatio)
	cacheHitRatioLock.Unlock()
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func CacheHitRatio2JSONString() string {
	cacheHitRatioLock.RLock()
	defer cacheHitRatioLock.RUnlock()
	jsonBytes, err := json.Marshal(CacheHitRatio)
	if err != nil {
		logger.SysError("error marshalling cache hit ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateCacheHitRatioByJSONString(jsonStr string) error {
	cacheHitRatioLock.Lock()
	defer cacheHitRatioLock.Unlock()
	CacheHitRatio = make(map[string]float64)
	if jsonStr == "" {
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), &CacheHitRatio)
}
