package controller

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/jhopan/Panway/common/helper"
	"github.com/jhopan/Panway/relay/constant/role"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common"
	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/common/logger"
	"github.com/jhopan/Panway/model"
	"github.com/jhopan/Panway/relay/adaptor/openai"
	billingratio "github.com/jhopan/Panway/relay/billing/ratio"
	"github.com/jhopan/Panway/relay/channeltype"
	"github.com/jhopan/Panway/relay/controller/validator"
	"github.com/jhopan/Panway/relay/meta"
	relaymodel "github.com/jhopan/Panway/relay/model"
	"github.com/jhopan/Panway/relay/relaymode"
)

func getAndValidateTextRequest(c *gin.Context, relayMode int) (*relaymodel.GeneralOpenAIRequest, error) {
	textRequest := &relaymodel.GeneralOpenAIRequest{}
	err := common.UnmarshalBodyReusable(c, textRequest)
	if err != nil {
		return nil, err
	}
	if relayMode == relaymode.Moderations && textRequest.Model == "" {
		textRequest.Model = "text-moderation-latest"
	}
	if relayMode == relaymode.Embeddings && textRequest.Model == "" {
		textRequest.Model = c.Param("model")
	}
	err = validator.ValidateTextRequest(textRequest, relayMode)
	if err != nil {
		return nil, err
	}
	return textRequest, nil
}

func getPromptTokens(textRequest *relaymodel.GeneralOpenAIRequest, relayMode int) int {
	switch relayMode {
	case relaymode.ChatCompletions:
		return openai.CountTokenMessages(textRequest.Messages, textRequest.Model)
	case relaymode.Completions:
		return openai.CountTokenInput(textRequest.Prompt, textRequest.Model)
	case relaymode.Moderations:
		return openai.CountTokenInput(textRequest.Input, textRequest.Model)
	}
	return 0
}

func getPreConsumedQuota(textRequest *relaymodel.GeneralOpenAIRequest, promptTokens int, ratio float64) int64 {
	preConsumedTokens := config.PreConsumedQuota + int64(promptTokens)
	if textRequest.MaxTokens != 0 {
		preConsumedTokens += int64(textRequest.MaxTokens)
	}
	return int64(float64(preConsumedTokens) * ratio)
}

func preConsumeQuota(ctx context.Context, textRequest *relaymodel.GeneralOpenAIRequest, promptTokens int, ratio float64, meta *meta.Meta) (int64, *relaymodel.ErrorWithStatusCode) {
	// Token rate limit per minute. Checked here because this is the first
	// place that knows the real prompt token count (the middleware only has
	// a rough character-based estimate).
	if limit := tokenTokenRateLimit(meta.TokenId); limit > 0 {
		// The upstream counts more tokens than the local estimate (it may add
		// its own system prompt), so the budget is checked against the tokens
		// actually consumed by the previous requests, and the real usage is
		// recorded after the request finishes.
		used := common.GetTokenRateLimiter().TokensUsed(meta.TokenId)
		if used >= limit {
			return 0, openai.ErrorWrapper(
				fmt.Errorf("rate limit exceeded: token reached %d tokens per minute", limit),
				"token_rate_limit_exceeded", http.StatusTooManyRequests)
		}
	}
	preConsumedQuota := getPreConsumedQuota(textRequest, promptTokens, ratio)

	userQuota, err := model.CacheGetUserQuota(ctx, meta.UserId)
	if err != nil {
		return preConsumedQuota, openai.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota-preConsumedQuota < 0 {
		return preConsumedQuota, openai.ErrorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}
	err = model.CacheDecreaseUserQuota(meta.UserId, preConsumedQuota)
	if err != nil {
		return preConsumedQuota, openai.ErrorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}
	if preConsumedQuota > 0 {
		// Always hold the token's own budget, even when the user balance is
		// large enough to be trusted. Skipping this step let a limited token
		// spend more than its remaining quota through concurrent requests.
		err := model.PreConsumeTokenQuota(meta.TokenId, preConsumedQuota)
		if err != nil {
			// The Redis copy of the user quota was already lowered; refresh
			// it from the database, which never lost any quota.
			if restoreErr := model.CacheUpdateUserQuota(ctx, meta.UserId); restoreErr != nil {
				logger.Error(ctx, "failed to restore user quota cache: "+restoreErr.Error())
			}
			return preConsumedQuota, openai.ErrorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}
	return preConsumedQuota, nil
}

func postConsumeQuota(ctx context.Context, usage *relaymodel.Usage, meta *meta.Meta, textRequest *relaymodel.GeneralOpenAIRequest, ratio float64, preConsumedQuota int64, modelRatio float64, groupRatio float64, systemPromptReset bool) {
	if usage == nil {
		logger.Error(ctx, "usage is nil, which is unexpected")
		return
	}
	var quota int64
	completionRatio := billingratio.GetCompletionRatio(textRequest.Model, meta.ChannelType)
	promptTokens := usage.PromptTokens
	completionTokens := usage.CompletionTokens
	cachedTokens := 0
	if usage.PromptTokensDetails != nil {
		cachedTokens = usage.PromptTokensDetails.CachedTokens
	}
	quota = computeQuota(promptTokens, cachedTokens, completionTokens, ratio, completionRatio, billingratio.GetCacheHitRatio(textRequest.Model))
	if ratio != 0 && quota <= 0 {
		quota = 1
	}
	// Record the real token usage for the per-token TPM limit. The local
	// estimate undercounts, so the upstream numbers are what count.
	if meta.TokenId != 0 {
		common.GetTokenRateLimiter().RecordTokens(meta.TokenId, promptTokens+completionTokens)
	}
	common.AddRelayUsage(int64(promptTokens+completionTokens), quota)
	totalTokens := promptTokens + completionTokens
	if totalTokens == 0 {
		// in this case, must be some error happened
		// we cannot just return, because we may have to return the pre-consumed quota
		quota = 0
	}
	quotaDelta := quota - preConsumedQuota
	err := model.PostConsumeTokenQuota(meta.TokenId, quotaDelta)
	if err != nil {
		logger.Error(ctx, "error consuming token remain quota: "+err.Error())
	}
	err = model.CacheUpdateUserQuota(ctx, meta.UserId)
	if err != nil {
		logger.Error(ctx, "error update user quota cache: "+err.Error())
	}
	logContent := fmt.Sprintf("Ratio: %.2f × %.2f × %.2f", modelRatio, groupRatio, completionRatio)
	if cachedTokens > 0 {
		logContent += fmt.Sprintf(" cache=%d", cachedTokens)
	}
	model.RecordConsumeLog(ctx, &model.Log{
		UserId:            meta.UserId,
		ChannelId:         meta.ChannelId,
		PromptTokens:      promptTokens,
		CachedTokens:      cachedTokens,
		CompletionTokens:  completionTokens,
		ModelName:         textRequest.Model,
		TokenName:         meta.TokenName,
		Quota:             int(quota),
		Content:           logContent,
		IsStream:          meta.IsStream,
		ElapsedTime:       helper.CalcElapsedTime(meta.StartTime),
		SystemPromptReset: systemPromptReset,
	})
	model.UpdateUserUsedQuotaAndRequestCount(meta.UserId, quota)
	model.UpdateChannelUsedQuota(meta.ChannelId, quota)
}

// computeQuota turns upstream token counts into quota units. Cached prompt
// tokens are billed with cacheRatio instead of the normal input price, so a
// prompt-cache hit is cheaper for the customer while Panway still records the
// real token counts.
func computeQuota(promptTokens, cachedTokens, completionTokens int, ratio, completionRatio, cacheRatio float64) int64 {
	if cachedTokens < 0 {
		cachedTokens = 0
	}
	if cachedTokens > promptTokens {
		cachedTokens = promptTokens
	}
	billablePrompt := float64(promptTokens-cachedTokens) + float64(cachedTokens)*cacheRatio
	return int64(math.Ceil((billablePrompt + float64(completionTokens)*completionRatio) * ratio))
}

func getMappedModelName(modelName string, mapping map[string]string) (string, bool) {
	if mapping == nil {
		return modelName, false
	}
	mappedModelName := mapping[modelName]
	if mappedModelName != "" {
		return mappedModelName, true
	}
	return modelName, false
}

func isErrorHappened(meta *meta.Meta, resp *http.Response) bool {
	if resp == nil {
		if meta.ChannelType == channeltype.AwsClaude {
			return false
		}
		return true
	}
	if resp.StatusCode != http.StatusOK &&
		// replicate return 201 to create a task
		resp.StatusCode != http.StatusCreated {
		return true
	}
	if meta.ChannelType == channeltype.DeepL {
		// skip stream check for deepl
		return false
	}

	if meta.IsStream && strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") &&
		// Even if stream mode is enabled, replicate will first return a task info in JSON format,
		// requiring the client to request the stream endpoint in the task info
		meta.ChannelType != channeltype.Replicate {
		return true
	}
	return false
}

func setSystemPrompt(ctx context.Context, request *relaymodel.GeneralOpenAIRequest, prompt string) (reset bool) {
	if prompt == "" {
		return false
	}
	if len(request.Messages) == 0 {
		return false
	}
	if request.Messages[0].Role == role.System {
		request.Messages[0].Content = prompt
		logger.Infof(ctx, "rewrite system prompt")
		return true
	}
	request.Messages = append([]relaymodel.Message{{
		Role:    role.System,
		Content: prompt,
	}}, request.Messages...)
	logger.Infof(ctx, "add system prompt")
	return true
}
