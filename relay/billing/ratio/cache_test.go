package ratio

import (
	"math"
	"testing"

	"github.com/jhopan/Panway/common/config"
)

func approx(got, want float64) bool {
	return math.Abs(got-want) < 1e-9
}

func resetCacheHitRatio() {
	cacheHitRatioLock.Lock()
	CacheHitRatio = make(map[string]float64)
	cacheHitRatioLock.Unlock()
}

func TestCacheHitRatioDefaultsToGlobalDiscount(t *testing.T) {
	resetCacheHitRatio()
	previous := config.CacheHitDiscountPercent
	config.CacheHitDiscountPercent = 80
	t.Cleanup(func() {
		config.CacheHitDiscountPercent = previous
		resetCacheHitRatio()
	})

	if got := GetCacheHitRatio("unknown/model"); !approx(got, 0.2) {
		t.Fatalf("default cache ratio harus 0.2, dapat %v", got)
	}
}

func TestSetAndDeleteCacheHitRatioOverride(t *testing.T) {
	resetCacheHitRatio()
	previous := config.CacheHitDiscountPercent
	config.CacheHitDiscountPercent = 80
	t.Cleanup(func() {
		config.CacheHitDiscountPercent = previous
		resetCacheHitRatio()
	})

	jsonStr, err := SetCacheHitRatio("agnes/agnes-3.0-flash", 0.5)
	if err != nil {
		t.Fatalf("gagal set override: %v", err)
	}
	if got := GetCacheHitRatio("agnes/agnes-3.0-flash"); !approx(got, 0.5) {
		t.Fatalf("override harus 0.5, dapat %v", got)
	}

	if err := UpdateCacheHitRatioByJSONString(jsonStr); err != nil {
		t.Fatalf("json map tidak bisa dibaca: %v", err)
	}
	if got := GetCacheHitRatio("agnes/agnes-3.0-flash"); !approx(got, 0.5) {
		t.Fatalf("setelah load map harus 0.5, dapat %v", got)
	}

	if _, err := DeleteCacheHitRatio("agnes/agnes-3.0-flash"); err != nil {
		t.Fatalf("gagal hapus override: %v", err)
	}
	if got := GetCacheHitRatio("agnes/agnes-3.0-flash"); !approx(got, 0.2) {
		t.Fatalf("setelah hapus harus kembali 0.2, dapat %v", got)
	}
}
