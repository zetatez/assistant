package psl

import (
	"context"
	"sync"

	"assistant/pkg/llm"
)

var (
	llmClient     llm.Client
	onceLLMClient sync.Once
)

func GetLLMClient() llm.Client { return llmClient }

func InitLLMClient() error {
	onceLLMClient.Do(func() {
		cfg := GetConfig().LLM
		if cfg.BaseURL == "" || cfg.APIKey == "" {
			GetLogger().Warn("llm: base_url or api_key is empty, LLM client disabled")
			return
		}
		llmClient = llm.NewClient(cfg)
		GetLogger().WithFields(map[string]interface{}{
			"base_url":      cfg.BaseURL,
			"text_models":   cfg.TextModels,
			"vision_models": cfg.VisionModels,
		}).Info("llm: client initialized")
	})
	return nil
}

func RegisterCleanupLLM() {
	RegisterCleanup(func(ctx context.Context) {
		llmClient = nil
	})
}
