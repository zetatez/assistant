package ollama

import (
	"testing"

	"assistant/pkg/llm"
)

func TestOllamaClient(t *testing.T) {
	client, err := New(llm.Config{
		BaseURL: "http://127.0.0.1:11434",
		Model:   "llama3.1:8b",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if client.Provider() != "ollama" {
		t.Errorf("expected provider 'ollama', got %s", client.Provider())
	}
	if client.Model() != "llama3.1:8b" {
		t.Errorf("expected model 'llama3.1:8b', got %s", client.Model())
	}

	caps := client.Capabilities()
	if !caps.Has(llm.CapabilityChat) {
		t.Error("ollama client should have chat capability")
	}
	if caps.Has(llm.CapabilityStream) {
		t.Error("ollama client should not have stream capability")
	}
}
