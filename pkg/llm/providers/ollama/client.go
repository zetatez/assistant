package ollama

// ollama 本地模型

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"assistant/pkg/llm"
)

type Client struct {
	baseURL string
	model   string
}

func init() {
	llm.Register("ollama", New)
}

func New(cfg llm.Config) (llm.Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://127.0.0.1:11434"
	}
	return &Client{
		baseURL: cfg.BaseURL,
		model:   cfg.Model, // e.g. llama3.1:8b
	}, nil
}

func (c *Client) Provider() string { return "ollama" }
func (c *Client) Model() string    { return c.model }

func (c *Client) Capabilities() llm.Capabilities {
	return llm.Capabilities{
		Supported: llm.CapabilityChat,
	}
}

func (c *Client) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	payload := map[string]any{
		"model":    c.model,
		"messages": req.Messages,
		"stream":   false,
	}

	b, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(
		ctx,
		"POST",
		c.baseURL+"/api/chat",
		bytes.NewReader(b),
	)
	httpReq.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	return &llm.ChatResponse{
		Content: raw.Message.Content,
	}, nil
}

func (c *Client) StreamChat(ctx context.Context, req llm.ChatRequest, cb llm.StreamCallback) error {
	return llm.ErrNotImplemented
}
