package controller

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/model"
	"github.com/jhopan/Panway/relay/billing/ratio"
)

// PricingEntry is one sellable model with its price.
type PricingEntry struct {
	// ModelId is the id a customer puts in the request body.
	ModelId string `json:"model_id"`
	// Provider is the part before the slash, e.g. "agnes" in
	// "agnes/agnes-3.0-flash". Empty when the id has no provider prefix.
	Provider string `json:"provider"`
	// Name is the model id without the provider prefix.
	Name string `json:"name"`
	// InputPrice is rupiah per one million input tokens.
	InputPrice float64 `json:"input_price"`
	// OutputPrice is rupiah per one million output tokens.
	OutputPrice float64 `json:"output_price"`
	// Currency of InputPrice and OutputPrice.
	Currency string `json:"currency"`
	// Unit explains the price unit, e.g. "per 1M tokens".
	Unit string `json:"unit"`
	// Requests7d and AvgLatencyMS come from the operation log, so the
	// catalogue can show how much a model is actually used.
	Requests7d   int64 `json:"requests_7d"`
	AvgLatencyMS int64 `json:"avg_latency_ms"`
}

// pricingCache holds the last answer of buildPricingEntries.
//
// /api/pricing is public and unauthenticated, so every landing page visit
// would otherwise run a channel query and walk the ratio maps. A price edit
// invalidates the cache directly; the TTL only bounds how long a channel or
// ratio change made from another page stays invisible.
var pricingCache struct {
	mu      sync.Mutex
	entries []PricingEntry
	expires time.Time
}

func invalidatePricingCache() {
	pricingCache.mu.Lock()
	pricingCache.entries = nil
	pricingCache.expires = time.Time{}
	pricingCache.mu.Unlock()
}

func cachedPricingEntries() ([]PricingEntry, error) {
	ttl := time.Duration(config.PricingCacheSeconds) * time.Second
	if ttl <= 0 {
		return buildPricingEntries()
	}
	pricingCache.mu.Lock()
	defer pricingCache.mu.Unlock()
	if pricingCache.entries != nil && time.Now().Before(pricingCache.expires) {
		return pricingCache.entries, nil
	}
	entries, err := buildPricingEntries()
	if err != nil {
		return nil, err
	}
	pricingCache.entries = entries
	pricingCache.expires = time.Now().Add(ttl)
	return entries, nil
}

// GetPricing returns the public price list.
//
// It is public on purpose: the landing page shows the price table before
// a customer signs up. The prices come from the live ratio maps, so the
// admin panel stays the single source of truth - editing Model Ratio or
// Completion Ratio in the panel changes what this endpoint returns.
//
// Quota math: quota = tokens * modelRatio * groupRatio, and QuotaPerUnit
// is 1 by default, which means one quota unit is one rupiah. So the price
// of one million tokens is modelRatio * 1e6 for input, and that value
// multiplied by the completion ratio for output.
func GetPricing(c *gin.Context) {
	entries, err := cachedPricingEntries()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "failed to load pricing",
			"data":    nil,
		})
		return
	}

	// The catalogue is public and changes rarely, so a short browser/proxy
	// cache keeps landing pages cheap. The admin console bypasses it with a
	// cache-busting query string when it edits a price.
	if config.PricingCacheSeconds > 0 {
		c.Header("Cache-Control", "public, max-age="+strconv.Itoa(config.PricingCacheSeconds))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"models":   entries,
			"currency": "IDR",
			"unit":     "1M tokens",
		},
	})
}

// buildPricingEntries reads the live channels and ratio maps and returns the
// sellable models. It is the uncached body of GetPricing.
func buildPricingEntries() ([]PricingEntry, error) {
	channels, err := model.GetAllChannels(0, 0, "all")
	if err != nil {
		return nil, err
	}

	groupRatio := ratio.GetGroupRatio("default")

	seen := make(map[string]bool)
	entries := make([]PricingEntry, 0)

	for _, ch := range channels {
		if ch.Status != model.ChannelStatusEnabled {
			continue
		}
		for _, modelId := range strings.Split(ch.Models, ",") {
			modelId = strings.TrimSpace(modelId)
			if modelId == "" || seen[modelId] {
				continue
			}
			seen[modelId] = true

			modelRatio := ratio.GetModelRatio(modelId, ch.Type)
			completionRatio := ratio.GetCompletionRatio(modelId, ch.Type)
			if modelRatio <= 0 {
				continue
			}

			inputPrice := modelRatio * groupRatio * 1000000
			outputPrice := inputPrice * completionRatio

			provider := ""
			name := modelId
			if idx := strings.Index(modelId, "/"); idx > 0 {
				provider = modelId[:idx]
				name = modelId[idx+1:]
			}

			entries = append(entries, PricingEntry{
				ModelId:     modelId,
				Provider:    provider,
				Name:        name,
				InputPrice:  inputPrice,
				OutputPrice: outputPrice,
				Currency:    "IDR",
				Unit:        "1M tokens",
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Provider != entries[j].Provider {
			return entries[i].Provider < entries[j].Provider
		}
		return entries[i].Name < entries[j].Name
	})

	type modelUsage struct {
		ModelName  string
		Requests   int64
		AvgLatency float64 `gorm:"column:avg_latency"`
	}
	var usage []modelUsage
	since := time.Now().Add(-7 * 24 * time.Hour).Unix()
	model.DB.Model(&model.Log{}).
		Select("model_name, count(*) as requests, avg(elapsed_time) as avg_latency").
		Where("type = ? AND created_at >= ?", model.LogTypeConsume, since).
		Group("model_name").
		Scan(&usage)
	byModel := make(map[string]modelUsage, len(usage))
	for _, row := range usage {
		byModel[row.ModelName] = row
	}
	for i := range entries {
		if stat, ok := byModel[entries[i].ModelId]; ok {
			entries[i].Requests7d = stat.Requests
			entries[i].AvgLatencyMS = int64(stat.AvgLatency)
		}
	}

	return entries, nil
}

// GetPriceHistory returns the price edits of the last 72 hours (or a custom
// window) so the catalogue can draw a small trend.
func GetPriceHistory(c *gin.Context) {
	since := time.Now().Add(-72 * time.Hour).Unix()
	if raw := c.Query("since"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed > 0 {
			since = parsed
		}
	}
	rows, err := model.GetPriceHistorySince(since)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "failed to load price history"})
		return
	}
	c.Header("Cache-Control", "public, max-age=30")
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": rows})
}
