package ratio

// PanRouter pricing.
//
// These are the models Panway sells. The upstream is PanRouter, which serves
// many providers behind one OpenAI-compatible surface, so the model id here
// is the full "provider/model" form (e.g. "agnes/agnes-3.0-flash").
//
// The numbers below are the upstream USD price per million tokens converted to
// the ratio unit one-api uses: quota = (prompt + completion * completionRatio)
// * modelRatio * groupRatio.
//
// Unit basis: one-api's own quota unit is "USD per 1K tokens / 0.002", so
// modelRatio is (usd_per_1M / 1000) / 0.002 = usd_per_1M / 2.
//
// With QuotaPerUnit set to 1 (1 quota = Rp1), the same numbers read as
// rupiah directly: multiply by the USD/IDR rate to get the rupiah price.
//
// Edit these through the admin panel (设置 -> 模型倍率) once the operator has
// settled the real sell price; the defaults here are a starting point.

// panRouterModelRatio maps a PanRouter model id to its input-price ratio.
var panRouterModelRatio = map[string]float64{
	// Agnes AI (free tier upstream, priced here as a paid model).
	"agnes/agnes-1.5-flash": 0.05,
	"agnes/agnes-2.0-flash": 0.05,
	"agnes/agnes-2.5-flash": 0.05,
	"agnes/agnes-3.0-flash": 0.05,

	// DeepSeek family.
	"th/deepseek-v4-flash:free": 0.07,
	"bzl/deepseek-v4-flash-0731free": 0.07,
	"cl/deepseek/deepseek-v4-flash": 0.07,
	"deepseek-v4-flash":         0.07,

	// Cloudflare Workers AI.
	"cf/@cf/qwen/qwen2.5-coder-32b-instruct": 0.05,
}

// panRouterCompletionRatio maps a PanRouter model id to output/input ratio.
var panRouterCompletionRatio = map[string]float64{
	"agnes/agnes-1.5-flash":                  4,
	"agnes/agnes-2.0-flash":                  4,
	"agnes/agnes-2.5-flash":                  4,
	"agnes/agnes-3.0-flash":                  4,
	"th/deepseek-v4-flash:free":              2,
	"bzl/deepseek-v4-flash-0731free":         2,
	"cl/deepseek/deepseek-v4-flash":          2,
	"deepseek-v4-flash":                      2,
	"cf/@cf/qwen/qwen2.5-coder-32b-instruct": 4,
}

// ApplyPanRouterPricing registers the PanRouter pricing into the live ratio
// maps. It is called once at startup so a fresh install already knows the
// price of every model Panway serves.
func ApplyPanRouterPricing() {
	for model, ratio := range panRouterModelRatio {
		if _, exists := ModelRatio[model]; !exists {
			ModelRatio[model] = ratio
		}
	}
	for model, ratio := range panRouterCompletionRatio {
		if _, exists := CompletionRatio[model]; !exists {
			CompletionRatio[model] = ratio
		}
	}
}
