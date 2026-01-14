package deepseek

import (
	"assistant/pkg/llm"
	"assistant/pkg/llm/providers/openai"
)

func init() {
	llm.Register("deepseek", New)
}

// New creates a DeepSeek client that reuses the OpenAI-compatible client.
// DeepSeek's API is compatible with the OpenAI API format.
func New(cfg llm.Config) (llm.Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.deepseek.com"
	}
	if cfg.Model == "" {
		cfg.Model = "deepseek-chat"
	}
	return openai.New(cfg)
}
