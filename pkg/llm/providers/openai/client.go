package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"assistant/pkg/llm"
)

type Client struct {
	apiKey  string
	baseURL string
	model   string
}

func init() {
	llm.Register("openai", New)
}

func New(cfg llm.Config) (llm.Client, error) {
	return &Client{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
	}, nil
}

func (c *Client) Provider() string { return "openai" }

func (c *Client) Model() string { return c.model }

func (c *Client) Capabilities() llm.Capabilities {
	return llm.Capabilities{
		Supported: llm.CapabilityChat |
			llm.CapabilityStream |
			llm.CapabilityFunctionCall,
	}
}

func (c *Client) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	payload := map[string]any{
		"model":       c.model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
	}

	b, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequestWithContext(
		ctx,
		"POST",
		c.baseURL+"/v1/chat/completions",
		bytes.NewReader(b),
	)

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw struct {
		Choices []struct {
			Message llm.Message `json:"message"`
		} `json:"choices"`
		Usage llm.TokenUsage `json:"usage"`
	}

	json.NewDecoder(resp.Body).Decode(&raw)

	return &llm.ChatResponse{
		Content: raw.Choices[0].Message.Content,
		Usage:   raw.Usage,
	}, nil
}

func (c *Client) StreamChat(ctx context.Context, req llm.ChatRequest, cb llm.StreamCallback) error {
	// 这里用 SSE，逻辑略
	return nil
}
