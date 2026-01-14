package deepseek

package deepseek

import (
	"context"

	"assistant/pkg/llm"
	"assistant/pkg/llm/providers/openai"
)

func init() {
	llm.Register("deepseek", New)
}

// DeepSeek 直接复用 OpenAI Client
func New(cfg llm.Config) (llm.Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.deepseek.com"
	}
	if cfg.Model == "" {
		cfg.Model = "deepseek-chat"
	}
	return openai.New(cfg)
}
