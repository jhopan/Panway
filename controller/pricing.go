package controller

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

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
	channels, err := model.GetAllChannels(0, 0, "all")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "failed to load pricing",
			"data":    nil,
		})
		return
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
