package openai

import (
	"context"
	"testing"

	"assistant/pkg/llm"
)

func TestOpenAIClient(t *testing.T) {
	// 注意：这是一个单元测试，不会实际调用 API
	// 我们主要测试客户端创建和能力声明
	client, err := New(llm.Config{
		APIKey:  "test-key",
		BaseURL: "https://api.openai.com",
		Model:   "gpt-4",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if client.Provider() != "openai" {
		t.Errorf("expected provider 'openai', got %s", client.Provider())
	}
	if client.Model() != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got %s", client.Model())
	}

	caps := client.Capabilities()
	if !caps.Has(llm.CapabilityChat) {
		t.Error("openai client should have chat capability")
	}
	if caps.Has(llm.CapabilityStream) {
		t.Error("openai client should not have stream capability (not implemented)")
	}
	if !caps.Has(llm.CapabilityFunctionCall) {
		t.Error("openai client should have function call capability")
	}
}
