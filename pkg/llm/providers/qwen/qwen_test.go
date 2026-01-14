package qwen

import (
	"testing"

	"assistant/pkg/llm"
)

func TestQwenClient(t *testing.T) {
	client, err := New(llm.Config{
		APIKey:  "test-key",
		BaseURL: "https://dashscope.aliyuncs.com",
		Model:   "qwen-plus",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if client.Provider() != "qwen" {
		t.Errorf("expected provider 'qwen', got %s", client.Provider())
	}
	if client.Model() != "qwen-plus" {
		t.Errorf("expected model 'qwen-plus', got %s", client.Model())
	}

	caps := client.Capabilities()
	if !caps.Has(llm.CapabilityChat) {
		t.Error("qwen client should have chat capability")
	}
	if caps.Has(llm.CapabilityStream) {
		t.Error("qwen client should not have stream capability")
	}
}
