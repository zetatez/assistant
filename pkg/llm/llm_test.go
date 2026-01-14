package llm

import (
	"context"
	"testing"
)

// MockClient 用于测试
type MockClient struct {
	provider     string
	model        string
	capabilities Capabilities
	chatImpl     func(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	streamImpl   func(ctx context.Context, req ChatRequest, cb StreamCallback) error
}

func (m *MockClient) Provider() string { return m.provider }
func (m *MockClient) Model() string    { return m.model }
func (m *MockClient) Capabilities() Capabilities {
	return m.capabilities
}
func (m *MockClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if m.chatImpl != nil {
		return m.chatImpl(ctx, req)
	}
	return nil, ErrNotImplemented
}
func (m *MockClient) StreamChat(ctx context.Context, req ChatRequest, cb StreamCallback) error {
	if m.streamImpl != nil {
		return m.streamImpl(ctx, req, cb)
	}
	return ErrNotImplemented
}

func TestChatWithoutCallback(t *testing.T) {
	mock := &MockClient{
		provider: "test",
		model:    "test-model",
		capabilities: Capabilities{
			Supported: CapabilityChat,
		},
		chatImpl: func(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
			return &ChatResponse{
				Content: "Hello, world!",
				Usage: TokenUsage{
					PromptTokens:     10,
					CompletionTokens: 5,
					TotalTokens:      15,
				},
			}, nil
		},
	}

	req := ChatRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Hello"},
		},
	}

	resp, err := Chat(context.Background(), mock, req, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Content != "Hello, world!" {
		t.Errorf("expected 'Hello, world!', got %s", resp.Content)
	}
	if resp.Usage.PromptTokens != 10 {
		t.Errorf("expected 10 prompt tokens, got %d", resp.Usage.PromptTokens)
	}
}

func TestChatWithCallbackButNoStreamCapability(t *testing.T) {
	mock := &MockClient{
		provider: "test",
		model:    "test-model",
		capabilities: Capabilities{
			Supported: CapabilityChat, // 没有流式能力
		},
		chatImpl: func(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
			return &ChatResponse{
				Content: "Fallback response",
			}, nil
		},
	}

	req := ChatRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Test"},
		},
	}

	called := false
	cb := func(delta string) {
		called = true
	}

	resp, err := Chat(context.Background(), mock, req, cb)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Content != "Fallback response" {
		t.Errorf("expected 'Fallback response', got %s", resp.Content)
	}
	if called {
		t.Error("callback should not have been called because client doesn't support streaming")
	}
}

func TestChatWithCallbackAndStreamCapability(t *testing.T) {
	mock := &MockClient{
		provider: "test",
		model:    "test-model",
		capabilities: Capabilities{
			Supported: CapabilityChat | CapabilityStream,
		},
		streamImpl: func(ctx context.Context, req ChatRequest, cb StreamCallback) error {
			cb("Hello")
			cb(" ")
			cb("World!")
			return nil
		},
	}

	req := ChatRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Stream test"},
		},
	}

	collected := ""
	cb := func(delta string) {
		collected += delta
	}

	resp, err := Chat(context.Background(), mock, req, cb)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != nil {
		t.Error("expected nil response when streaming succeeds")
	}
	if collected != "Hello World!" {
		t.Errorf("expected 'Hello World!', got %s", collected)
	}
}

func TestChatWithCallbackAndStreamNotImplemented(t *testing.T) {
	mock := &MockClient{
		provider: "test",
		model:    "test-model",
		capabilities: Capabilities{
			Supported: CapabilityChat | CapabilityStream,
		},
		streamImpl: func(ctx context.Context, req ChatRequest, cb StreamCallback) error {
			return ErrNotImplemented
		},
		chatImpl: func(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
			return &ChatResponse{
				Content: "Fallback chat response",
			}, nil
		},
	}

	req := ChatRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Test"},
		},
	}

	streamCalled := false
	cb := func(delta string) {
		streamCalled = true
	}

	resp, err := Chat(context.Background(), mock, req, cb)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Content != "Fallback chat response" {
		t.Errorf("expected 'Fallback chat response', got %s", resp.Content)
	}
	if streamCalled {
		t.Error("callback should not have been called because StreamChat returned ErrNotImplemented")
	}
}

func TestChatNoCapability(t *testing.T) {
	mock := &MockClient{
		provider: "test",
		model:    "test-model",
		capabilities: Capabilities{
			Supported: 0, // 没有任何能力
		},
	}

	req := ChatRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Test"},
		},
	}

	_, err := Chat(context.Background(), mock, req, nil)
	if err != ErrNotImplemented {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}

func TestCapabilitiesHas(t *testing.T) {
	caps := Capabilities{
		Supported: CapabilityChat | CapabilityStream,
	}

	if !caps.Has(CapabilityChat) {
		t.Error("should have chat capability")
	}
	if !caps.Has(CapabilityStream) {
		t.Error("should have stream capability")
	}
	if caps.Has(CapabilityFunctionCall) {
		t.Error("should not have function call capability")
	}
	if caps.Has(CapabilityVision) {
		t.Error("should not have vision capability")
	}
}
