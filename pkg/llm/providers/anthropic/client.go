package anthropic

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
	llm.Register("anthropic", New)
}

func New(cfg llm.Config) (llm.Client, error) {
	return &Client{
		apiKey:  cfg.APIKey,
		baseURL: "https://api.anthropic.com",
		model:   cfg.Model, // e.g. claude-3-5-sonnet-20241022
	}, nil
}

func (c *Client) Provider() string { return "anthropic" }
func (c *Client) Model() string    { return c.model }

func (c *Client) Capabilities() llm.Capabilities {
	return llm.Capabilities{
		Supported: llm.CapabilityChat,
	}
}

func (c *Client) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	// Anthropic 把 system 单独拎出来
	var system string
	var messages []map[string]string

	for _, m := range req.Messages {
		if m.Role == llm.RoleSystem {
			system = m.Content
			continue
		}
		messages = append(messages, map[string]string{
			"role":    string(m.Role),
			"content": m.Content,
		})
	}

	payload := map[string]any{
		"model":    c.model,
		"system":   system,
		"messages": messages,
		"max_tokens": func() int {
			if req.MaxTokens > 0 {
				return req.MaxTokens
			}
			return 1024
		}(),
	}

	b, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(
		ctx,
		"POST",
		c.baseURL+"/v1/messages",
		bytes.NewReader(b),
	)

	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	return &llm.ChatResponse{
		Content: raw.Content[0].Text,
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
