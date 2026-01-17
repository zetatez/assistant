package qwen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"assistant/pkg/llm"
)

type Client struct {
	apiKey  string
	baseURL string
	model   string
	client  *llm.BaseClient
}

func init() {
	llm.Register("qwen", New)
}

func New(cfg llm.Config) (llm.Client, error) {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com"
	}

	model := cfg.Model
	if model == "" {
		model = "qwen-plus"
	}

	return &Client{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   model,
		client:  llm.NewBaseClient(baseURL, cfg),
	}, nil
}

func (c *Client) Provider() string { return "qwen" }

func (c *Client) Model() string { return c.model }

func (c *Client) Capabilities() llm.Capabilities {
	return llm.Capabilities{Supported: llm.CapabilityChat | llm.CapabilityStream}
}

func (c *Client) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	payload := map[string]any{
		"model": c.getModel(req.Model),
		"input": map[string]any{
			"messages": req.Messages,
		},
		"parameters": map[string]any{
			"temperature": req.Temperature,
		},
	}

	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
	}

	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
		"Content-Type":  "application/json",
	}

	resp, err := c.client.Do(ctx, "POST", "/api/v1/services/aigc/text-generation/generation", payload, headers)
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
		return nil, &llm.ProviderError{Code: raw.Code, Message: raw.Message}
	}

	if len(raw.Output.Choices) == 0 {
		return nil, llm.ErrMaxRetries
	}

	return &llm.ChatResponse{
		Content: raw.Output.Choices[0].Message.Content,
		Role:    raw.Output.Choices[0].Message.Role,
		Usage: llm.TokenUsage{
			PromptTokens:     raw.Usage.InputTokens,
			CompletionTokens: raw.Usage.OutputTokens,
			TotalTokens:      raw.Usage.InputTokens + raw.Usage.OutputTokens,
		},
	}, nil
}

func (c *Client) StreamChat(ctx context.Context, req llm.ChatRequest, cb llm.StreamCallback) error {
	params := map[string]any{
		"temperature":        req.Temperature,
		"incremental_output": true,
	}
	if req.MaxTokens > 0 {
		params["max_tokens"] = req.MaxTokens
	}

	payload := map[string]any{
		"model": c.getModel(req.Model),
		"input": map[string]any{
			"messages": req.Messages,
		},
		"parameters": params,
	}

	headers := map[string]string{
		"Authorization":   "Bearer " + c.apiKey,
		"Content-Type":    "application/json",
		"Accept":          "text/event-stream",
		"X-DashScope-SSE": "enable",
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/services/aigc/text-generation/generation", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.client.HTTPClient().Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return &llm.HTTPError{Code: resp.StatusCode, Message: strings.TrimSpace(string(b))}
	}

	err = llm.ReadSSE(ctx, resp.Body, func(data string) error {
		if data == "[DONE]" {
			return io.EOF
		}

		var raw struct {
			Output struct {
				Choices []struct {
					Message      llm.Message `json:"message"`
					FinishReason *string     `json:"finish_reason"`
				} `json:"choices"`
			} `json:"output"`
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
			Code    string `json:"code"`
			Message string `json:"message"`
		}

		if err := json.Unmarshal([]byte(data), &raw); err != nil {
			return err
		}
		if raw.Code != "" {
			return &llm.ProviderError{Code: raw.Code, Message: raw.Message}
		}
		if len(raw.Output.Choices) == 0 {
			return nil
		}

		cb(llm.ChatResponse{
			Content: raw.Output.Choices[0].Message.Content,
			Role:    raw.Output.Choices[0].Message.Role,
			Usage: llm.TokenUsage{
				PromptTokens:     raw.Usage.InputTokens,
				CompletionTokens: raw.Usage.OutputTokens,
				TotalTokens:      raw.Usage.InputTokens + raw.Usage.OutputTokens,
			},
		})
		if raw.Output.Choices[0].FinishReason != nil && *raw.Output.Choices[0].FinishReason != "" {
			return io.EOF
		}
		return nil
	})
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("qwen stream: %w", err)
	}
	return nil
}

func (c *Client) getModel(model string) string {
	if model != "" {
		return model
	}
	return c.model
}
