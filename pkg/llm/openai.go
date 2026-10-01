package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Config 描述单个 OpenAI 兼容供应商的配置。
// 模型列表顺序即优先级：优先使用列表靠前的模型，失败时依次向后回退。
type Config struct {
	BaseURL      string   `mapstructure:"base_url"`
	APIKey       string   `mapstructure:"api_key"`
	TextModels   []string `mapstructure:"text_models"`
	VisionModels []string `mapstructure:"vision_models"`
	Timeout      int      `mapstructure:"timeout"`
	Temperature  float32  `mapstructure:"temperature"`
}

// OpenAIClient 是一个面向 OpenAI 兼容接口的单供应商客户端。
type OpenAIClient struct {
	cfg    Config
	client *http.Client
}

func NewClient(cfg Config) *OpenAIClient {
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}
	if cfg.Timeout > 0 {
		client.Timeout = time.Duration(cfg.Timeout) * time.Second
	}
	return &OpenAIClient{cfg: cfg, client: client}
}

func NewClientWithHTTP(cfg Config, httpClient *http.Client) *OpenAIClient {
	c := NewClient(cfg)
	if httpClient != nil {
		c.client = httpClient
	}
	return c
}

// Config 返回供应商配置。
func (c *OpenAIClient) Config() Config { return c.cfg }

// Chat 调用上游聊天补全接口。模型按列表顺序回退：
//   - req.Model 非空时作为显式模型直接使用；
//   - 否则若消息含图片则按 VisionModels 顺序，否则按 TextModels 顺序逐一尝试，直到成功或全部失败。
func (c *OpenAIClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	models := c.pickModels(req)
	if len(models) == 0 {
		return nil, fmt.Errorf("llm: no model configured")
	}

	var lastErr error
	for _, model := range models {
		resp, err := c.doChat(ctx, model, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("llm: all models failed: %w", lastErr)
}

func (c *OpenAIClient) pickModels(req ChatRequest) []string {
	if req.Model != "" {
		return []string{req.Model}
	}
	if hasImageInMessages(req.Messages) && len(c.cfg.VisionModels) > 0 {
		return c.cfg.VisionModels
	}
	return c.cfg.TextModels
}

func (c *OpenAIClient) doChat(ctx context.Context, model string, req ChatRequest) (*ChatResponse, error) {
	payload, err := c.buildPayload(model, req)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	target := strings.TrimRight(c.cfg.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "assistant/1.0")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		return nil, &HTTPError{Code: resp.StatusCode, Message: providerErrorMessage(raw)}
	}

	var raw struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
		Usage TokenUsage `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("llm: decode response: %w", err)
	}
	if len(raw.Choices) == 0 {
		return nil, fmt.Errorf("llm: empty choices")
	}
	return &ChatResponse{
		Content:   raw.Choices[0].Message.Content,
		Role:      raw.Choices[0].Message.Role,
		ToolCalls: raw.Choices[0].Message.ToolCalls,
		Usage:     raw.Usage,
	}, nil
}

func (c *OpenAIClient) buildPayload(model string, req ChatRequest) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"model":    model,
		"messages": convertMessages(req.Messages),
		"stream":   false,
	}
	temperature := req.Temperature
	if temperature == 0 {
		temperature = c.cfg.Temperature
	}
	if temperature != 0 {
		payload["temperature"] = temperature
	}
	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
	}
	if len(req.Tools) > 0 {
		tools := make([]map[string]interface{}, 0, len(req.Tools))
		for _, t := range req.Tools {
			tools = append(tools, map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        t.Name,
					"description": t.Description,
					"parameters":  t.Parameters,
				},
			})
		}
		payload["tools"] = tools
	}
	return payload, nil
}

type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string { return e.Message }

type imageURLContent struct {
	URL string `json:"url"`
}

type messageContent struct {
	Type     string           `json:"type"`
	Text     string           `json:"text,omitempty"`
	ImageURL *imageURLContent `json:"image_url,omitempty"`
}

func convertMessages(msgs []Message) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(msgs))
	for _, m := range msgs {
		if m.ImageBase64 != "" {
			content := []messageContent{
				{Type: "text", Text: m.Content},
				{Type: "image_url", ImageURL: &imageURLContent{URL: "data:image/jpeg;base64," + m.ImageBase64}},
			}
			result = append(result, map[string]interface{}{
				"role":    m.Role,
				"content": content,
			})
			continue
		}

		msg := map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		}
		if m.ToolCallID != "" {
			msg["tool_call_id"] = m.ToolCallID
		}
		if len(m.ToolCalls) > 0 {
			msg["tool_calls"] = m.ToolCalls
		}
		result = append(result, msg)
	}
	return result
}

func hasImageInMessages(msgs []Message) bool {
	for _, m := range msgs {
		if m.ImageBase64 != "" {
			return true
		}
	}
	return false
}

func providerErrorMessage(body []byte) string {
	if len(body) == 0 {
		return "provider failed"
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return string(body)
	}
	if msg, ok := parsed["error"].(string); ok {
		return msg
	}
	if errObj, ok := parsed["error"].(map[string]interface{}); ok {
		if msg, ok := errObj["message"].(string); ok {
			return msg
		}
	}
	return string(body)
}
