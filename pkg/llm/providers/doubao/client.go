package doubao

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"assistant/pkg/llm"
)

type Client struct {
	apiKey     string
	baseURL    string
	endpointID string
}

func init() {
	llm.Register("doubao", New)
}

func New(cfg llm.Config) (llm.Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://ark.cn-beijing.volces.com"
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("doubao requires endpoint_id in Config.Model")
	}
	return &Client{
		apiKey:     cfg.APIKey,
		baseURL:    cfg.BaseURL,
		endpointID: cfg.Model,
	}, nil
}

func (c *Client) Provider() string { return "doubao" }

func (c *Client) Model() string { return c.endpointID }

func (c *Client) Capabilities() llm.Capabilities {
	return llm.Capabilities{
		Supported: llm.CapabilityChat,
	}
}

func (c *Client) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	payload := map[string]any{
		"messages":    req.Messages,
		"temperature": req.Temperature,
	}

	b, _ := json.Marshal(payload)

	url := fmt.Sprintf(
		"%s/api/v3/chat/completions/%s",
		c.baseURL,
		c.endpointID,
	)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
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
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	if raw.Error != nil {
		return nil, fmt.Errorf("doubao error: %s", raw.Error.Message)
	}

	return &llm.ChatResponse{
		Content: raw.Choices[0].Message.Content,
		Usage:   raw.Usage,
	}, nil
}

func (c *Client) StreamChat(ctx context.Context, req llm.ChatRequest, cb llm.StreamCallback) error {
	return llm.ErrNotImplemented
}
