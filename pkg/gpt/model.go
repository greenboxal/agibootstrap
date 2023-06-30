package gpt

import (
	"os"
	"path"
	"strings"

	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"
)

var GlobalClient = createNewClient()
var GlobalEmbedder = &openai.Embedder{
	Client: GlobalClient,
	Model:  openai.AdaEmbeddingV2,
}
var GlobalModel = &openai.ChatLanguageModel{
	Client:      GlobalClient,
	Model:       "gpt-3.5-turbo-16k",
	Temperature: 1.0,
}

var GlobalModelTokenizer = tokenizers.TikTokenForModel(GlobalModel.Model)

func createNewClient() *openai.Client {
	if os.Getenv("OPENAI_API_KEY") == "" {
		home := os.Getenv("HOME")

		if home != "" {
			p := path.Join(home, ".openai", "api-key")
			key, err := os.ReadFile(p)

			if err == nil {
				_ = os.Setenv("OPENAI_API_KEY", strings.TrimSpace(string(key)))
			}
		}
	}

	return openai.NewClient()
}
