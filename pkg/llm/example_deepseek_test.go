package llm

import (
	"context"
	"os"
	"testing"
)

// TestDeepSeekExample demonstrates how to use the DeepSeek provider with a mock client.
// This test does not make real API calls.
func TestDeepSeekExample(t *testing.T) {
	// Create a mock client that simulates DeepSeek's behavior
	mock := &MockClient{
		provider: "deepseek",
		model:    "deepseek-chat",
		capabilities: Capabilities{
			Supported: CapabilityChat | CapabilityStream,
		},
		chatImpl: func(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
			// Simulate a typical DeepSeek response
			return &ChatResponse{
				Content: "Hello! I'm DeepSeek, an AI assistant created by DeepSeek Company.",
				Usage: TokenUsage{
					PromptTokens:     15,
					CompletionTokens: 20,
					TotalTokens:      35,
				},
			}, nil
		},
	}

	// Prepare a chat request
	req := ChatRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: RoleSystem, Content: "You are a helpful assistant."},
			{Role: RoleUser, Content: "Hello, who are you?"},
		},
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	// Make the chat call without streaming
	resp, err := Chat(context.Background(), mock, req, nil)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	// Verify the response
	expectedContent := "Hello! I'm DeepSeek, an AI assistant created by DeepSeek Company."
	if resp.Content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, resp.Content)
	}
	if resp.Usage.TotalTokens != 35 {
		t.Errorf("Expected 35 total tokens, got %d", resp.Usage.TotalTokens)
	}

	// Test streaming with callback
	collected := ""
	streamCb := func(delta string) {
		collected += delta
	}

	mock.streamImpl = func(ctx context.Context, req ChatRequest, cb StreamCallback) error {
		cb("Hello! ")
		cb("I'm ")
		cb("DeepSeek.")
		return nil
	}

	// This should use streaming since we provide a callback and client supports it
	resp, err = Chat(context.Background(), mock, req, streamCb)
	if err != nil {
		t.Fatalf("Streaming chat failed: %v", err)
	}
	// When streaming succeeds, resp should be nil
	if resp != nil {
		t.Error("Expected nil response when streaming succeeds")
	}
	if collected != "Hello! I'm DeepSeek." {
		t.Errorf("Expected collected %q, got %q", "Hello! I'm DeepSeek.", collected)
	}
}

// TestDeepSeekRealIntegration demonstrates actual usage of the DeepSeek provider.
// This test is skipped unless the DEEPSEEK_API_KEY environment variable is set.
func TestDeepSeekRealIntegration(t *testing.T) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: DEEPSEEK_API_KEY not set")
	}

	// Create a real DeepSeek client
	cfg := Config{
		APIKey:  apiKey,
		BaseURL: "https://api.deepseek.com", // DeepSeek's API endpoint
		Model:   "deepseek-chat",
	}

	client, err := NewClient("deepseek", cfg)
	if err != nil {
		t.Fatalf("Failed to create DeepSeek client: %v", err)
	}

	// Verify the client's capabilities
	caps := client.Capabilities()
	if !caps.Has(CapabilityChat) {
		t.Error("DeepSeek client should support chat capability")
	}
	// Note: DeepSeek may or may not support streaming via the OpenAI-compatible API
	// We'll just log it for information
	t.Logf("DeepSeek client capabilities: %v", caps.Supported)

	// Prepare a simple request
	req := ChatRequest{
		Model: cfg.Model,
		Messages: []Message{
			{Role: RoleUser, Content: "Say 'Hello, World!'"},
		},
		Temperature: 0.7,
		MaxTokens:   50,
	}

	// Make the API call (without streaming)
	ctx := context.Background()
	resp, err := Chat(ctx, client, req, nil)
	if err != nil {
		t.Fatalf("DeepSeek API call failed: %v", err)
	}

	t.Logf("DeepSeek response: %s", resp.Content)
	t.Logf("Token usage: prompt=%d, completion=%d, total=%d",
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)

	// Basic validation
	if resp.Content == "" {
		t.Error("Expected non-empty response content")
	}
}

// TestDeepSeekRegistration verifies that the DeepSeek provider is properly registered.
func TestDeepSeekRegistration(t *testing.T) {
	// The provider should be registered via its init() function
	// We can test by trying to create a client with minimal config
	cfg := Config{
		APIKey:  "test-key",
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-chat",
	}

	client, err := NewClient("deepseek", cfg)
	if err != nil {
		t.Fatalf("Failed to create DeepSeek client (maybe not registered): %v", err)
	}

	// Check that it's the right provider
	if client.Provider() != "deepseek" {
		t.Errorf("Expected provider 'deepseek', got %q", client.Provider())
	}
	if client.Model() != "deepseek-chat" {
		t.Errorf("Expected model 'deepseek-chat', got %q", client.Model())
	}
}
