package common

import (
	"fmt"
	"sync/atomic"
)

// Lightweight process counters exposed at /metrics.
//
// Panway already runs without Prometheus; adding a client library just for
// four counters would grow the binary and the dependency tree. The text
// exposition format below is understood by Prometheus and Grafana Agent.
var (
	relayRequests  atomic.Int64
	relayErrors    atomic.Int64
	relayTokens    atomic.Int64
	relayQuotaUsed atomic.Int64
)

func IncRelayRequest() {
	relayRequests.Add(1)
}

func IncRelayError() {
	relayErrors.Add(1)
}

func AddRelayUsage(tokens, quota int64) {
	if tokens > 0 {
		relayTokens.Add(tokens)
	}
	if quota > 0 {
		relayQuotaUsed.Add(quota)
	}
}

// MetricsText renders the counters in Prometheus text exposition format.
func MetricsText() string {
	return fmt.Sprintf(`# HELP panway_relay_requests_total Relay requests received.
# TYPE panway_relay_requests_total counter
panway_relay_requests_total %d
# HELP panway_relay_errors_total Relay requests that ended with an error.
# TYPE panway_relay_errors_total counter
panway_relay_errors_total %d
# HELP panway_relay_tokens_total Prompt and completion tokens billed since start.
# TYPE panway_relay_tokens_total counter
panway_relay_tokens_total %d
# HELP panway_quota_billed_total Quota billed since start.
# TYPE panway_quota_billed_total counter
panway_quota_billed_total %d
`, relayRequests.Load(), relayErrors.Load(), relayTokens.Load(), relayQuotaUsed.Load())
}
