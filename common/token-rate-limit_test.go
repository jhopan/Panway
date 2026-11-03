package common

import (
	"sync"
	"testing"
	"time"
)

func TestAllowRequestRespectsRPM(t *testing.T) {
	l := &TokenRateLimiter{
		requests: make(map[int][]int64),
		tokens:   make(map[int][]tokenBucket),
		inflight: make(map[int]int),
		window:   time.Minute,
	}
	if !l.AllowRequest(1, 3) {
		t.Fatal("request 1 harus boleh")
	}
	if !l.AllowRequest(1, 3) {
		t.Fatal("request 2 harus boleh")
	}
	if !l.AllowRequest(1, 3) {
		t.Fatal("request 3 harus boleh")
	}
	if l.AllowRequest(1, 3) {
		t.Fatal("request 4 harus ditolak")
	}
	// Token lain tidak terpengaruh.
	if !l.AllowRequest(2, 3) {
		t.Fatal("token lain harus punya kuota sendiri")
	}
}

func TestAllowRequestUnlimitedWhenZero(t *testing.T) {
	l := &TokenRateLimiter{
		requests: make(map[int][]int64),
		tokens:   make(map[int][]tokenBucket),
		inflight: make(map[int]int),
		window:   time.Minute,
	}
	for i := 0; i < 100; i++ {
		if !l.AllowRequest(1, 0) {
			t.Fatalf("request %d harus boleh saat limit 0", i)
		}
	}
}

func TestAllowRequestSlidingWindow(t *testing.T) {
	l := &TokenRateLimiter{
		requests: make(map[int][]int64),
		tokens:   make(map[int][]tokenBucket),
		inflight: make(map[int]int),
		window:   100 * time.Millisecond,
	}
	l.AllowRequest(1, 1)
	if l.AllowRequest(1, 1) {
		t.Fatal("kedua request dalam window harus ditolak")
	}
	time.Sleep(150 * time.Millisecond)
	if !l.AllowRequest(1, 1) {
		t.Fatal("setelah window lewat, request harus boleh")
	}
}

func TestAllowTokensRespectsTPM(t *testing.T) {
	l := &TokenRateLimiter{
		requests: make(map[int][]int64),
		tokens:   make(map[int][]tokenBucket),
		inflight: make(map[int]int),
		window:   time.Minute,
	}
	if !l.AllowTokens(1, 100, 1000) {
		t.Fatal("100 token harus boleh")
	}
	if l.AllowTokens(1, 950, 1000) {
		t.Fatal("950 token lagi harus ditolak (total 1050 > 1000)")
	}
	if !l.AllowTokens(1, 900, 1000) {
		t.Fatal("900 token lagi harus boleh (total 1000)")
	}
	if l.AllowTokens(1, 1, 1000) {
		t.Fatal("1 token lagi harus ditolak (total 1001)")
	}
}

func TestAllowTokensUnlimitedWhenZero(t *testing.T) {
	l := &TokenRateLimiter{
		requests: make(map[int][]int64),
		tokens:   make(map[int][]tokenBucket),
		inflight: make(map[int]int),
		window:   time.Minute,
	}
	if !l.AllowTokens(1, 1_000_000, 0) {
		t.Fatal("limit 0 harus mengizinkan semua")
	}
}

func TestConcurrencyLimit(t *testing.T) {
	l := &TokenRateLimiter{
		requests: make(map[int][]int64),
		tokens:   make(map[int][]tokenBucket),
		inflight: make(map[int]int),
		window:   time.Minute,
	}
	release1, ok := l.Acquire(1, 2)
	if !ok {
		t.Fatal("acquire pertama harus boleh")
	}
	_, ok = l.Acquire(1, 2)
	if !ok {
		t.Fatal("acquire kedua harus boleh")
	}
	_, ok = l.Acquire(1, 2)
	if ok {
		t.Fatal("acquire ketiga harus ditolak")
	}
	release1()
	release, ok := l.Acquire(1, 2)
	if !ok {
		t.Fatal("setelah release, acquire harus boleh")
	}
	release()
}

func TestConcurrencyUnlimitedWhenZero(t *testing.T) {
	l := &TokenRateLimiter{
		requests: make(map[int][]int64),
		tokens:   make(map[int][]tokenBucket),
		inflight: make(map[int]int),
		window:   time.Minute,
	}
	for i := 0; i < 50; i++ {
		if _, ok := l.Acquire(1, 0); !ok {
			t.Fatalf("acquire %d harus boleh saat limit 0", i)
		}
	}
}

func TestReleaseIsIdempotent(t *testing.T) {
	l := &TokenRateLimiter{
		requests: make(map[int][]int64),
		tokens:   make(map[int][]tokenBucket),
		inflight: make(map[int]int),
		window:   time.Minute,
	}
	release, _ := l.Acquire(1, 1)
	release()
	release() // kedua kali tidak boleh mengurangi lagi
	_, ok := l.Acquire(1, 1)
	if !ok {
		t.Fatal("release ganda tidak boleh menahan slot")
	}
}

func TestConcurrentUse(t *testing.T) {
	l := &TokenRateLimiter{
		requests: make(map[int][]int64),
		tokens:   make(map[int][]tokenBucket),
		inflight: make(map[int]int),
		window:   time.Minute,
	}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.AllowRequest(1, 1000)
			l.AllowTokens(1, 10, 100000)
			release, ok := l.Acquire(1, 100)
			if ok {
				release()
			}
		}()
	}
	wg.Wait()
}

func TestRecordTokensThenTokensUsed(t *testing.T) {
	l := &TokenRateLimiter{
		requests: make(map[int][]int64),
		tokens:   make(map[int][]tokenBucket),
		inflight: make(map[int]int),
		window:   time.Minute,
	}
	l.RecordTokens(1, 497)
	l.RecordTokens(1, 503)
	if got := l.TokensUsed(1); got != 1000 {
		t.Fatalf("want 1000 tokens used, got %d", got)
	}
	if got := l.TokensUsed(2); got != 0 {
		t.Fatalf("token lain harus 0, got %d", got)
	}
}

func TestTokensUsedDrivesAllowTokens(t *testing.T) {
	l := &TokenRateLimiter{
		requests: make(map[int][]int64),
		tokens:   make(map[int][]tokenBucket),
		inflight: make(map[int]int),
		window:   time.Minute,
	}
	// Usage nyata dari request sebelumnya: 497 token terpakai, limit 600.
	l.RecordTokens(1, 497)
	if used := l.TokensUsed(1); used < 600 {
		// Request berikutnya harus ditolak karena sisa budget tinggal 103.
		if l.AllowTokens(1, 497, 600) {
			t.Fatalf("request berikutnya harus ditolak: used=%d limit=600", used)
		}
	}
}
