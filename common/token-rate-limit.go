package common

import (
	"sync"
	"time"
)

// TokenRateLimiter tracks per-token request rate (RPM) and token rate (TPM),
// plus the number of requests in flight (concurrency).
//
// It is a sliding window: every accepted request keeps its timestamp for the
// length of the window, and the oldest entry is dropped once it falls out of
// the window. A fixed window would let a client burst twice the limit across
// the window boundary.
//
// The limiter is safe for concurrent use.
type TokenRateLimiter struct {
	mu sync.Mutex
	// requests holds the timestamps (UnixNano) of accepted requests, per token id.
	requests map[int][]int64
	// tokens holds the token counts of accepted requests, per token id, with
	// the second each entry belongs to.
	tokens map[int][]tokenBucket
	// inflight counts requests currently running, per token id.
	inflight map[int]int
	// window is the length of the sliding window in seconds.
	window time.Duration
	// lastSweep is when expired entries were last removed.
	lastSweep time.Time
}

type tokenBucket struct {
	second int64
	count  int
}

var tokenLimiter = &TokenRateLimiter{
	requests: make(map[int][]int64),
	tokens:   make(map[int][]tokenBucket),
	inflight: make(map[int]int),
	window:   time.Minute,
}

// GetTokenRateLimiter returns the shared per-token rate limiter.
func GetTokenRateLimiter() *TokenRateLimiter {
	return tokenLimiter
}

// SetWindow sets the sliding window length. Used by tests.
func (l *TokenRateLimiter) SetWindow(d time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.window = d
}

// AllowRequest checks and consumes one request slot for a token id.
// It returns false when the request rate limit is exceeded.
func (l *TokenRateLimiter) AllowRequest(tokenId, maxRPM int) bool {
	if maxRPM <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sweepLocked()

	now := time.Now().UnixNano()
	queue := l.requests[tokenId]
	// Drop entries that fell out of the window.
	cutoff := now - int64(l.window)
	kept := queue[:0]
	for _, ts := range queue {
		if ts > cutoff {
			kept = append(kept, ts)
		}
	}
	if len(kept) >= maxRPM {
		l.requests[tokenId] = kept
		return false
	}
	l.requests[tokenId] = append(kept, now)
	return true
}

// TokensUsed returns the tokens consumed by a token id inside the window.
func (l *TokenRateLimiter) TokensUsed(tokenId int) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sweepLocked()
	return l.tokensUsedLocked(tokenId, time.Now().Unix())
}

func (l *TokenRateLimiter) tokensUsedLocked(tokenId int, now int64) int {
	used := 0
	for _, b := range l.tokens[tokenId] {
		if now-b.second < int64(l.window.Seconds()) {
			used += b.count
		}
	}
	return used
}

// RecordTokens adds consumed tokens to the token id's budget.
//
// The upstream counts more tokens than the local estimate (an upstream may
// add its own system prompt), so the real usage is recorded after the request
// rather than an estimate beforehand.
func (l *TokenRateLimiter) RecordTokens(tokenId, tokens int) {
	if tokens <= 0 {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sweepLocked()

	now := time.Now().Unix()
	buckets := l.tokens[tokenId]
	if len(buckets) > 0 && buckets[len(buckets)-1].second == now {
		buckets[len(buckets)-1].count += tokens
	} else {
		buckets = append(buckets, tokenBucket{second: now, count: tokens})
	}
	l.tokens[tokenId] = buckets
}

// AllowTokens checks whether adding tokens stays within the budget, and
// records them when it does. It returns false when the limit is exceeded.
func (l *TokenRateLimiter) AllowTokens(tokenId, tokens, maxTPM int) bool {
	if maxTPM <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sweepLocked()

	now := time.Now().Unix()
	if l.tokensUsedLocked(tokenId, now)+tokens > maxTPM {
		return false
	}
	buckets := l.tokens[tokenId]
	if len(buckets) > 0 && buckets[len(buckets)-1].second == now {
		buckets[len(buckets)-1].count += tokens
	} else {
		buckets = append(buckets, tokenBucket{second: now, count: tokens})
	}
	l.tokens[tokenId] = buckets
	return true
}

// Acquire starts a request for a token id, enforcing the concurrency limit.
// The returned release function must be called when the request finishes.
func (l *TokenRateLimiter) Acquire(tokenId, maxConcurrency int) (func(), bool) {
	if maxConcurrency <= 0 {
		return func() {}, true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.inflight[tokenId] >= maxConcurrency {
		return nil, false
	}
	l.inflight[tokenId]++
	released := false
	return func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		if released {
			return
		}
		released = true
		if l.inflight[tokenId] > 0 {
			l.inflight[tokenId]--
		}
	}, true
}

// sweepLocked removes expired entries. The caller must hold the lock.
func (l *TokenRateLimiter) sweepLocked() {
	now := time.Now()
	if now.Sub(l.lastSweep) < l.window {
		return
	}
	l.lastSweep = now
	cutoff := now.Add(-l.window).UnixNano()
	for id, queue := range l.requests {
		kept := queue[:0]
		for _, ts := range queue {
			if ts > cutoff {
				kept = append(kept, ts)
			}
		}
		if len(kept) == 0 {
			delete(l.requests, id)
		} else {
			l.requests[id] = kept
		}
	}
	for id, buckets := range l.tokens {
		kept := buckets[:0]
		for _, b := range buckets {
			if b.second > cutoff {
				kept = append(kept, b)
			}
		}
		if len(kept) == 0 {
			delete(l.tokens, id)
		} else {
			l.tokens[id] = kept
		}
	}
	for id, count := range l.inflight {
		if count <= 0 {
			delete(l.inflight, id)
		}
	}
}
