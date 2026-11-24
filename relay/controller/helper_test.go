package controller

import "testing"

func TestComputeQuotaChargesCachedPromptCheaper(t *testing.T) {
	// 1000 prompt tokens at ratio 0.05, completion ratio 4, cache ratio 0.2.
	full := computeQuota(1000, 0, 100, 0.05, 4, 0.2)
	cached := computeQuota(1000, 1000, 100, 0.05, 4, 0.2)

	if full != 70 {
		t.Fatalf("tagihan normal harus 70, dapat %d", full)
	}
	if cached != 30 {
		t.Fatalf("cache hit harus 30 (20%% dari input), dapat %d", cached)
	}
	if cached >= full {
		t.Fatalf("cache hit harus lebih murah: full %d cached %d", full, cached)
	}
}

func TestComputeQuotaClampsCachedTokens(t *testing.T) {
	got := computeQuota(100, 500, 0, 0.05, 4, 0.2)
	if got != 1 {
		t.Fatalf("cache tidak boleh melebihi prompt; hasil %d", got)
	}
}
