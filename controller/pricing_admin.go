package controller

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/model"
	"github.com/jhopan/Panway/relay/billing/ratio"
)

// PricingUpdate is the admin form for changing one model's price.
//
// The operator edits rupiah per one million tokens, not the internal ratio
// unit, because that is the number they actually sell at.
type PricingUpdate struct {
	// ModelId is the model to change, e.g. "agnes/agnes-3.0-flash".
	//
	// It travels in the body rather than the path because model ids
	// contain slashes, which a path parameter cannot hold.
	ModelId string `json:"model_id"`
	// InputPrice is rupiah per one million input tokens.
	InputPrice float64 `json:"input_price"`
	// OutputPrice is rupiah per one million output tokens.
	OutputPrice float64 `json:"output_price"`
}

// UpdatePricing changes the price of one model.
//
// It writes the same ModelRatio / CompletionRatio options the JSON editor
// in the settings page writes, so there is still exactly one source of
// truth for pricing.
//
// Conversion: the quota unit is 1 rupiah by default (QuotaPerUnit = 1),
// and quota = tokens * modelRatio * groupRatio, so the ratio for one
// million tokens is price / 1e6 / groupRatio. The completion ratio is the
// output price divided by the input price.
func UpdatePricing(c *gin.Context) {
	var update PricingUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "invalid request body",
		})
		return
	}

	modelId := strings.TrimSpace(update.ModelId)
	if modelId == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "the model id is required",
		})
		return
	}

	if update.InputPrice <= 0 || update.OutputPrice <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "prices must be greater than zero",
		})
		return
	}

	groupRatio := ratio.GetGroupRatio("default")
	if groupRatio <= 0 {
		groupRatio = 1
	}

	modelRatio := update.InputPrice / 1000000 / groupRatio
	completionRatio := update.OutputPrice / update.InputPrice

	modelRatioJSON, err := ratio.SetModelRatio(modelId, modelRatio)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	completionRatioJSON, err := ratio.SetCompletionRatio(modelId, completionRatio)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Persist both maps so the change survives a restart.
	if err := model.UpdateOption("ModelRatio", modelRatioJSON); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if err := model.UpdateOption("CompletionRatio", completionRatioJSON); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"model_id":     modelId,
			"input_price":  update.InputPrice,
			"output_price": update.OutputPrice,
		},
	})
}

// DeletePricing removes a model from the price list.
//
// The model is only removed from the ratio maps; the channel still decides
// which models are actually served.
func DeletePricing(c *gin.Context) {
	var update PricingUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "invalid request body",
		})
		return
	}

	modelId := strings.TrimSpace(update.ModelId)
	if modelId == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "the model id is required",
		})
		return
	}

	modelRatioJSON, completionRatioJSON, err := ratio.DeleteModelRatio(modelId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if err := model.UpdateOption("ModelRatio", modelRatioJSON); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if err := model.UpdateOption("CompletionRatio", completionRatioJSON); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

// ListPricedModels returns the models that are actually served, with the
// price the operator is selling them at.
//
// Only models that a live channel offers are listed. The ratio maps also
// carry every default model One API knows about, and those are not
// something this deployment can serve, so listing them would only bury
// the real catalogue.
func ListPricedModels(c *gin.Context) {
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
	models := make([]ratio.PricedModel, 0)

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

			inputPrice := 0.0
			if modelRatio > 0 {
				inputPrice = modelRatio * groupRatio * 1000000
			}
			outputPrice := 0.0
			if inputPrice > 0 {
				outputPrice = inputPrice * completionRatio
			}

			models = append(models, ratio.PricedModel{
				ModelId:         modelId,
				InputPrice:      inputPrice,
				OutputPrice:     outputPrice,
				CompletionRatio: completionRatio,
			})
		}
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].ModelId < models[j].ModelId
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    models,
	})
}
