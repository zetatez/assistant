package psl

import (
	"context"
	"sync"

	"assistant/pkg/llmproxy"
)

var (
	llmProxySvc   *llmproxy.ProxyService
	llmClient     llmproxy.Client
	onceLLMClient sync.Once
)

func GetProxyService() *llmproxy.ProxyService { return llmProxySvc }

func GetLLMClient() llmproxy.Client { return llmClient }

func InitLLMClient() error {
	onceLLMClient.Do(func() {
		cfg := GetConfig().LLMProxy
		cfg.VPN = GetConfig().Settings.VPN
		llmProxySvc = llmproxy.NewProxyService(cfg)
		if llmProxySvc.HasProviders() {
			llmClient = llmproxy.NewProxyClient(llmProxySvc)
		}
	})
	return nil
}

func RegisterCleanupLLM() {
	RegisterCleanup(func(ctx context.Context) {
		llmClient = nil
		llmProxySvc = nil
	})
}
