package doubao

import (
	"testing"

	"assistant/pkg/llm"
)

func TestDoubaoClient(t *testing.T) {
	client, err := New(llm.Config{
		APIKey: "test-key",
		Model:  "endpoint-123", // endpoint_id
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if client.Provider() != "doubao" {
		t.Errorf("expected provider 'doubao', got %s", client.Provider())
	}
	if client.Model() != "endpoint-123" {
		t.Errorf("expected model 'endpoint-123', got %s", client.Model())
	}

	caps := client.Capabilities()
	if !caps.Has(llm.CapabilityChat) {
		t.Error("doubao client should have chat capability")
	}
	if caps.Has(llm.CapabilityStream) {
		t.Error("doubao client should not have stream capability")
	}
}

func TestDoubaoClientRequiresModel(t *testing.T) {
	_, err := New(llm.Config{
		APIKey: "test-key",
		Model:  "", // 空模型应该报错
	})
	if err == nil {
		t.Error("expected error when model is empty")
	}
}
