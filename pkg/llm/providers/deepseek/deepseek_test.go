package deepseek

import (
	"testing"

	"assistant/pkg/llm"
)

func TestDeepSeekClient(t *testing.T) {
	client, err := New(llm.Config{
		APIKey:  "test-key",
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-chat",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// DeepSeek 使用 openai 客户端，所以 provider 应该是 "openai"
	if client.Provider() != "openai" {
		t.Errorf("expected provider 'openai', got %s", client.Provider())
	}
	if client.Model() != "deepseek-chat" {
		t.Errorf("expected model 'deepseek-chat', got %s", client.Model())
	}
}
