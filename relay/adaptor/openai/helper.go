package openai

import (
	"fmt"
	"strings"
	"sync"

	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/relay/channeltype"
	"github.com/jhopan/Panway/relay/model"
)

func ResponseText2Usage(responseText string, modelName string, promptTokens int) *model.Usage {
	usage := &model.Usage{}
	usage.PromptTokens = promptTokens
	usage.CompletionTokens = CountTokenText(responseText, modelName)
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	return usage
}

// promptOverheadByModel remembers how many prompt tokens an upstream adds on
// top of the local count.
//
// PanRouter injects a hidden system prompt and prefixes the request with it:
// the same tiny message that counts 10 tokens locally is reported by the
// upstream as 499. When the upstream cannot measure a stream it sends an
// estimate instead, so the known overhead from earlier exact usage is added
// back to keep stream billing on par with non-stream billing.
var promptOverhead = struct {
	sync.Mutex
	byModel map[string]int
}{byModel: map[string]int{}}

func learnPromptOverhead(modelName string, localPromptTokens, upstreamPromptTokens int) {
	overhead := upstreamPromptTokens - localPromptTokens
	if modelName == "" || overhead <= 0 {
		return
	}
	promptOverhead.Lock()
	promptOverhead.byModel[modelName] = overhead
	promptOverhead.Unlock()
}

func learnedPromptOverhead(modelName string) int {
	promptOverhead.Lock()
	defer promptOverhead.Unlock()
	return promptOverhead.byModel[modelName]
}

// NormalizeEstimatedUsage replaces an upstream estimate with Panway's own
// token counts.
//
// Some upstreams cannot measure usage while streaming and send an estimate
// instead (flagged with "estimated": true). The estimate can be several times
// the real prompt size, because it reserves a fixed budget for the upstream's
// hidden system prompt. Billing that number would overcharge the customer, so
// an estimated usage is rebuilt from the request prompt, the known hidden
// prompt overhead, and the streamed response text. Exact usage is returned
// untouched, and ESTIMATED_USAGE_MODE=upstream turns the whole normalization
// off for operators who want to bill the upstream meter as-is.
func NormalizeEstimatedUsage(usage *model.Usage, responseText string, modelName string, promptTokens int) *model.Usage {
	if usage == nil || !usage.Estimated {
		return usage
	}
	if config.EstimatedUsageMode == "upstream" {
		return usage
	}
	normalized := *usage
	if promptTokens > 0 {
		normalized.PromptTokens = promptTokens + learnedPromptOverhead(modelName)
	}
	completionTokens := CountTokenText(responseText, modelName)
	if completionTokens == 0 && responseText == "" {
		// Tool calls and reasoning-only answers carry no text but still use
		// output tokens; keep the upstream count rather than billing zero.
		completionTokens = usage.CompletionTokens
	}
	normalized.CompletionTokens = completionTokens
	normalized.TotalTokens = normalized.PromptTokens + completionTokens
	normalized.Estimated = false
	return &normalized
}

func GetFullRequestURL(baseURL string, requestURL string, channelType int) string {
	if channelType == channeltype.OpenAICompatible {
		return fmt.Sprintf("%s%s", strings.TrimSuffix(baseURL, "/"), strings.TrimPrefix(requestURL, "/v1"))
	}
	fullRequestURL := fmt.Sprintf("%s%s", baseURL, requestURL)

	if strings.HasPrefix(baseURL, "https://gateway.ai.cloudflare.com") {
		switch channelType {
		case channeltype.OpenAI:
			fullRequestURL = fmt.Sprintf("%s%s", baseURL, strings.TrimPrefix(requestURL, "/v1"))
		case channeltype.Azure:
			fullRequestURL = fmt.Sprintf("%s%s", baseURL, strings.TrimPrefix(requestURL, "/openai/deployments"))
		}
	}
	return fullRequestURL
}
