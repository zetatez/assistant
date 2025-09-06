package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"assistant/pkg/llm"
)

type Client struct {
	apiKey  string
	baseURL string
	model   string
}

func init() {
	llm.Register("deepseek", New)
}

func New(cfg llm.Config) (llm.Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("deepseek: API key is required")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}

	model := cfg.Model
	if model == "" {
		model = "deepseek-chat"
	}

	return &Client{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   model,
	}, nil
}

func (c *Client) Provider() string { return "deepseek" }

func (c *Client) Model() string { return c.model }

func (c *Client) Capabilities() llm.Capabilities {
	return llm.Capabilities{
		Supported: llm.CapabilityChat,
	}
}

func (c *Client) Chat(
	ctx context.Context,
	req llm.ChatRequest,
) (*llm.ChatResponse, error) {
	payload := map[string]any{
		"model":       c.model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		c.baseURL+"/v1/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deepseek: http %d", resp.StatusCode)
	}

	var raw struct {
		Choices []struct {
			Message llm.Message `json:"message"`
		} `json:"choices"`
		Usage llm.TokenUsage `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	if len(raw.Choices) == 0 {
		return nil, fmt.Errorf("deepseek: empty response")
	}

	return &llm.ChatResponse{
		Content: raw.Choices[0].Message.Content,
		Usage:   raw.Usage,
	}, nil
}

func (c *Client) StreamChat(
	ctx context.Context,
	req llm.ChatRequest,
	cb llm.StreamCallback,
) error {
	return llm.ErrNotImplemented
}
