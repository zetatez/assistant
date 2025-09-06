package anthropic

import (
	"testing"

	"assistant/pkg/llm"
)

func TestAnthropicClient(t *testing.T) {
	client, err := New(llm.Config{
		APIKey: "test-key",
		Model:  "claude-3-5-sonnet-20241022",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if client.Provider() != "anthropic" {
		t.Errorf("expected provider 'anthropic', got %s", client.Provider())
	}
	if client.Model() != "claude-3-5-sonnet-20241022" {
		t.Errorf("expected model 'claude-3-5-sonnet-20241022', got %s", client.Model())
	}

	caps := client.Capabilities()
	if !caps.Has(llm.CapabilityChat) {
		t.Error("anthropic client should have chat capability")
	}
	if caps.Has(llm.CapabilityStream) {
		t.Error("anthropic client should not have stream capability")
	}
}
