package aiproxy

import "github.com/jhopan/Panway/relay/adaptor/openai"

var ModelList = []string{""}

func init() {
	ModelList = openai.ModelList
}
