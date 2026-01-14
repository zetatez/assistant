package llm

import (
	"testing"
)

func TestRegisterAndNewClient(t *testing.T) {
	// 保存原始 providers
	originalProviders := make(map[string]Factory)
	for k, v := range providers {
		originalProviders[k] = v
	}
	defer func() {
		providers = originalProviders
	}()

	// 清空 providers 进行测试
	providers = make(map[string]Factory)

	// 注册一个测试工厂
	testFactory := func(cfg Config) (Client, error) {
		return &mockClientImpl{
			provider: "test-provider",
			model:    cfg.Model,
		}, nil
	}

	Register("test", testFactory)

	// 测试 NewClient 可以创建客户端
	client, err := NewClient("test", Config{Model: "test-model"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client.Provider() != "test-provider" {
		t.Errorf("expected provider 'test-provider', got %s", client.Provider())
	}
	if client.Model() != "test-model" {
		t.Errorf("expected model 'test-model', got %s", client.Model())
	}

	// 测试未知 provider
	_, err = NewClient("unknown", Config{})
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

// mockClientImpl 用于测试
type mockClientImpl struct {
	provider string
	model    string
}

func (m *mockClientImpl) Provider() string { return m.provider }
func (m *mockClientImpl) Model() string    { return m.model }
func (m *mockClientImpl) Capabilities() Capabilities {
	return Capabilities{
		Supported: CapabilityChat,
	}
}
func (m *mockClientImpl) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	return nil, ErrNotImplemented
}
func (m *mockClientImpl) StreamChat(ctx context.Context, req ChatRequest, cb StreamCallback) error {
	return ErrNotImplemented
}
