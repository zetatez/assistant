package doubao

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

type imageURLContent struct {
	URL string `json:"url"`
}

type messageContent struct {
	Type     string           `json:"type"`
	Text     string           `json:"text,omitempty"`
	ImageURL *imageURLContent `json:"image_url,omitempty"`
}

type Client struct {
	apiKey     string
	baseURL    string
	endpointID string
	httpClient *http.Client
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
		httpClient: cfg.GetHTTPClient(),
	}, nil
}

func (c *Client) Provider() string { return "doubao" }

func (c *Client) Model() string { return c.endpointID }

func (c *Client) Capabilities() llm.Capabilities {
	return llm.Capabilities{
		Supported: llm.CapabilityChat | llm.CapabilityStream,
	}
}

func (c *Client) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	messages := convertMessages(req.Messages)

	payload := map[string]any{
		"messages":    messages,
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

	resp, err := c.httpClient.Do(httpReq)
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
	payload := map[string]any{
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"stream":      true,
	}
	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
	}
	if req.TopP > 0 {
		payload["top_p"] = req.TopP
	}
	if len(req.Tools) > 0 {
		payload["tools"] = req.Tools
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf(
		"%s/api/v3/chat/completions/%s",
		c.baseURL,
		c.endpointID,
	)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return &llm.HTTPError{Code: resp.StatusCode, Message: strings.TrimSpace(string(body))}
	}

	err = llm.ReadSSE(ctx, resp.Body, func(data string) error {
		if data == "[DONE]" {
			return io.EOF
		}

		var raw struct {
			Choices []struct {
				Delta struct {
					Content   string         `json:"content"`
					Role      llm.Role       `json:"role"`
					ToolCalls []llm.ToolCall `json:"tool_calls"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
			Usage llm.TokenUsage `json:"usage"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(data), &raw); err != nil {
			return err
		}
		if raw.Error != nil {
			return &llm.ProviderError{Message: raw.Error.Message}
		}
		if len(raw.Choices) == 0 {
			return nil
		}

		cb(llm.ChatResponse{
			Content:   raw.Choices[0].Delta.Content,
			Role:      raw.Choices[0].Delta.Role,
			ToolCalls: raw.Choices[0].Delta.ToolCalls,
			Usage:     raw.Usage,
		})
		if raw.Choices[0].FinishReason != nil && *raw.Choices[0].FinishReason != "" {
			return io.EOF
		}
		return nil
	})
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("doubao stream: %w", err)
	}
	return nil
}

func convertMessages(msgs []llm.Message) []map[string]any {
	result := make([]map[string]any, 0, len(msgs))
	for _, m := range msgs {
		if m.ImageBase64 != "" {
			content := []messageContent{
				{Type: "text", Text: m.Content},
				{Type: "image_url", ImageURL: &imageURLContent{URL: "data:image/jpeg;base64," + m.ImageBase64}},
			}
			result = append(result, map[string]any{
				"role":    m.Role,
				"content": content,
			})
		} else {
			result = append(result, map[string]any{
				"role":    m.Role,
				"content": m.Content,
			})
		}
	}
	return result
}
