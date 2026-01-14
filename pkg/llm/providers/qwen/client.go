package qwen

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
	llm.Register("qwen", New)
}

func New(cfg llm.Config) (llm.Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://dashscope.aliyuncs.com"
	}
	if cfg.Model == "" {
		cfg.Model = "qwen-plus"
	}
	return &Client{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
	}, nil
}

func (c *Client) Provider() string { return "qwen" }

func (c *Client) Model() string    { return c.model }

func (c *Client) Capabilities() llm.Capabilities {
	return llm.Capabilities{
		Supported: llm.CapabilityChat,
	}
}

func (c *Client) StreamChat(...) error {
	return llm.ErrNotImplemented
}

func (c *Client) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	payload := map[string]any{
		"model": c.model,
		"input": map[string]any{
			"messages": req.Messages,
		},
		"parameters": map[string]any{
			"temperature": req.Temperature,
		},
	}

	b, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/api/v1/services/aigc/text-generation/generation", c.baseURL),
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
		Output struct {
			Choices []struct {
				Message llm.Message `json:"message"`
			} `json:"choices"`
		} `json:"output"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	if raw.Code != "" {
		return nil, fmt.Errorf("qwen error: %s", raw.Message)
	}

	return &llm.ChatResponse{
		Content: raw.Output.Choices[0].Message.Content,
		Usage: llm.TokenUsage{
			PromptTokens:     raw.Usage.InputTokens,
			CompletionTokens: raw.Usage.OutputTokens,
			TotalTokens:      raw.Usage.InputTokens + raw.Usage.OutputTokens,
		},
	}, nil
}

func (c *Client) StreamChat(ctx context.Context, req llm.ChatRequest, cb llm.StreamCallback) error {
	return llm.ErrNotImplemented
}
