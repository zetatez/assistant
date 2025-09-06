package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"assistant/pkg/llm"
)

type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

func init() {
	llm.Register("ollama", New)
}

func New(cfg llm.Config) (llm.Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://127.0.0.1:11434"
	}
	return &Client{
		baseURL:    cfg.BaseURL,
		model:      cfg.Model, // e.g. llama3.1:8b
		httpClient: cfg.GetHTTPClient(),
	}, nil
}

func (c *Client) Provider() string { return "ollama" }
func (c *Client) Model() string    { return c.model }

func (c *Client) Capabilities() llm.Capabilities {
	return llm.Capabilities{
		Supported: llm.CapabilityChat | llm.CapabilityStream,
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

	resp, err := c.httpClient.Do(httpReq)
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
	payload := map[string]any{
		"model":    c.model,
		"messages": req.Messages,
		"stream":   true,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		c.baseURL+"/api/chat",
		bytes.NewReader(b),
	)
	if err != nil {
		return err
	}
	httpReq.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return &llm.HTTPError{Code: resp.StatusCode, Message: string(body)}
	}

	dec := json.NewDecoder(resp.Body)
	for {
		var raw struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			Done  bool   `json:"done"`
			Error string `json:"error"`
		}

		if err := dec.Decode(&raw); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if raw.Error != "" {
			return fmt.Errorf("ollama error: %s", raw.Error)
		}

		if raw.Message.Content != "" || raw.Message.Role != "" {
			cb(llm.ChatResponse{Content: raw.Message.Content, Role: llm.Role(raw.Message.Role)})
		}
		if raw.Done {
			return nil
		}
	}
}
