package relay

import (
	"github.com/jhopan/Panway/relay/adaptor"
	"github.com/jhopan/Panway/relay/adaptor/aiproxy"
	"github.com/jhopan/Panway/relay/adaptor/ali"
	"github.com/jhopan/Panway/relay/adaptor/anthropic"
	"github.com/jhopan/Panway/relay/adaptor/aws"
	"github.com/jhopan/Panway/relay/adaptor/baidu"
	"github.com/jhopan/Panway/relay/adaptor/cloudflare"
	"github.com/jhopan/Panway/relay/adaptor/cohere"
	"github.com/jhopan/Panway/relay/adaptor/coze"
	"github.com/jhopan/Panway/relay/adaptor/deepl"
	"github.com/jhopan/Panway/relay/adaptor/gemini"
	"github.com/jhopan/Panway/relay/adaptor/ollama"
	"github.com/jhopan/Panway/relay/adaptor/openai"
	"github.com/jhopan/Panway/relay/adaptor/palm"
	"github.com/jhopan/Panway/relay/adaptor/proxy"
	"github.com/jhopan/Panway/relay/adaptor/replicate"
	"github.com/jhopan/Panway/relay/adaptor/tencent"
	"github.com/jhopan/Panway/relay/adaptor/vertexai"
	"github.com/jhopan/Panway/relay/adaptor/xunfei"
	"github.com/jhopan/Panway/relay/adaptor/zhipu"
	"github.com/jhopan/Panway/relay/apitype"
)

func GetAdaptor(apiType int) adaptor.Adaptor {
	switch apiType {
	case apitype.AIProxyLibrary:
		return &aiproxy.Adaptor{}
	case apitype.Ali:
		return &ali.Adaptor{}
	case apitype.Anthropic:
		return &anthropic.Adaptor{}
	case apitype.AwsClaude:
		return &aws.Adaptor{}
	case apitype.Baidu:
		return &baidu.Adaptor{}
	case apitype.Gemini:
		return &gemini.Adaptor{}
	case apitype.OpenAI:
		return &openai.Adaptor{}
	case apitype.PaLM:
		return &palm.Adaptor{}
	case apitype.Tencent:
		return &tencent.Adaptor{}
	case apitype.Xunfei:
		return &xunfei.Adaptor{}
	case apitype.Zhipu:
		return &zhipu.Adaptor{}
	case apitype.Ollama:
		return &ollama.Adaptor{}
	case apitype.Coze:
		return &coze.Adaptor{}
	case apitype.Cohere:
		return &cohere.Adaptor{}
	case apitype.Cloudflare:
		return &cloudflare.Adaptor{}
	case apitype.DeepL:
		return &deepl.Adaptor{}
	case apitype.VertexAI:
		return &vertexai.Adaptor{}
	case apitype.Proxy:
		return &proxy.Adaptor{}
	case apitype.Replicate:
		return &replicate.Adaptor{}
	}
	return nil
}
