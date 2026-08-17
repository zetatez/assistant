package llmproxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIReqToAnthropic converts an OpenAI /chat/completions request body
// to Anthropic /v1/messages format.
func OpenAIReqToAnthropic(body []byte) ([]byte, error) {
	var req struct {
		Model       string                   `json:"model"`
		Messages    []map[string]interface{} `json:"messages"`
		MaxTokens   int                      `json:"max_tokens,omitempty"`
		Temperature float64                  `json:"temperature,omitempty"`
		Stream      bool                     `json:"stream,omitempty"`
		Tools       []map[string]interface{} `json:"tools,omitempty"`
		TopP        float64                  `json:"top_p,omitempty"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	var system strings.Builder
	var msgs []map[string]interface{}

	for _, m := range req.Messages {
		role, _ := m["role"].(string)
		content, _ := m["content"].(string)

		switch role {
		case "system":
			if content != "" {
				if system.Len() > 0 {
					system.WriteString("\n")
				}
				system.WriteString(content)
			}
			continue
		case "assistant":
			am := map[string]interface{}{"role": role}
			if tcs, ok := m["tool_calls"].([]interface{}); ok && len(tcs) > 0 {
				var blocks []map[string]interface{}
				for _, raw := range tcs {
					call, _ := raw.(map[string]interface{})
					fn, _ := call["function"].(map[string]interface{})
					blocks = append(blocks, map[string]interface{}{
						"type":  "tool_use",
						"id":    call["id"],
						"name":  fn["name"],
						"input": json.RawMessage(fmt.Sprintf("%s", fn["arguments"])),
					})
				}
				am["content"] = blocks
			} else {
				am["content"] = content
			}
			msgs = append(msgs, am)
			continue
		case "tool":
			toolCallID, _ := m["tool_call_id"].(string)
			msgs = append(msgs, map[string]interface{}{
				"role": "user",
				"content": []map[string]interface{}{{
					"type":        "tool_result",
					"tool_use_id": toolCallID,
					"content":     content,
				}},
			})
			continue
		default: // user
			msgs = append(msgs, map[string]interface{}{"role": role, "content": content})
		}
	}

	out := map[string]interface{}{
		"model":    req.Model,
		"messages": msgs,
		"stream":   req.Stream,
	}
	if system.Len() > 0 {
		out["system"] = system.String()
	}
	if req.MaxTokens > 0 {
		out["max_tokens"] = req.MaxTokens
	}
	if req.Temperature != 0 {
		out["temperature"] = req.Temperature
	}
	if req.TopP != 0 {
		out["top_p"] = req.TopP
	}
	if len(req.Tools) > 0 {
		out["tools"] = openAIToolsToAnthropic(req.Tools)
	}
	return json.Marshal(out)
}

func openAIToolsToAnthropic(tools []map[string]interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	for _, t := range tools {
		fn, _ := t["function"].(map[string]interface{})
		if fn == nil {
			continue
		}
		tool := map[string]interface{}{
			"name":         fn["name"],
			"description":  fn["description"],
			"input_schema": fn["parameters"],
		}
		out = append(out, tool)
	}
	return out
}

// AnthropicRespToOpenAI converts an Anthropic /v1/messages response body
// to OpenAI /chat/completions format.
func AnthropicRespToOpenAI(body []byte) ([]byte, error) {
	var res struct {
		ID         string `json:"id"`
		Model      string `json:"model"`
		StopReason string `json:"stop_reason"`
		Content    []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage *struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	var text strings.Builder
	for _, c := range res.Content {
		if c.Type == "text" {
			text.WriteString(c.Text)
		}
	}

	finish := "stop"
	if res.StopReason == "max_tokens" {
		finish = "length"
	} else if res.StopReason == "tool_use" {
		finish = "tool_calls"
	}

	out := map[string]interface{}{
		"id":      "chatcmpl-" + strings.TrimPrefix(res.ID, "msg_"),
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   res.Model,
		"choices": []map[string]interface{}{{
			"index":         0,
			"message":       map[string]interface{}{"role": "assistant", "content": text.String()},
			"finish_reason": finish,
		}},
	}
	if res.Usage != nil {
		out["usage"] = map[string]int{
			"prompt_tokens":     res.Usage.InputTokens,
			"completion_tokens": res.Usage.OutputTokens,
			"total_tokens":      res.Usage.InputTokens + res.Usage.OutputTokens,
		}
	}
	return json.Marshal(out)
}

// AnthropicStreamToOpenAI converts an Anthropic SSE stream to an OpenAI
// chat.completion.chunk SSE stream.
func AnthropicStreamToOpenAI(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	id := "chatcmpl-stream"
	model := ""
	created := time.Now().Unix()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" {
			continue
		}
		var ev struct {
			Type    string `json:"type"`
			Message struct {
				ID    string `json:"id"`
				Model string `json:"model"`
			} `json:"message"`
			Delta struct {
				Type       string `json:"type"`
				Text       string `json:"text"`
				StopReason string `json:"stop_reason"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			continue
		}

		switch ev.Type {
		case "message_start":
			if ev.Message.ID != "" {
				id = "chatcmpl-" + strings.TrimPrefix(ev.Message.ID, "msg_")
			}
			if ev.Message.Model != "" {
				model = ev.Message.Model
			}
			writeChunk(w, id, model, created, map[string]interface{}{"role": "assistant", "content": ""}, nil)
		case "content_block_delta":
			if ev.Delta.Type == "text_delta" && ev.Delta.Text != "" {
				writeChunk(w, id, model, created, map[string]interface{}{"content": ev.Delta.Text}, nil)
			}
		case "message_delta":
			finish := "stop"
			if ev.Delta.StopReason == "max_tokens" {
				finish = "length"
			} else if ev.Delta.StopReason == "tool_use" {
				finish = "tool_calls"
			}
			writeChunk(w, id, model, created, map[string]interface{}{}, &finish)
		case "message_stop":
			_, _ = fmt.Fprintf(w, "data: [DONE]\n\n")
			return nil
		}
	}
	return scanner.Err()
}

func writeChunk(w io.Writer, id, model string, created int64, delta map[string]interface{}, finish *string) {
	chunk := map[string]interface{}{
		"id":      id,
		"object":  "chat.completion.chunk",
		"created": created,
		"model":   model,
		"choices": []map[string]interface{}{{
			"index":         0,
			"delta":         delta,
			"finish_reason": finish,
		}},
	}
	b, _ := json.Marshal(chunk)
	_, _ = fmt.Fprintf(w, "data: %s\n\n", b)
}

// convertAnthropicResponse wraps an Anthropic upstream response into an
// OpenAI-format http.Response (handles both streaming and non-streaming).
func convertAnthropicResponse(resp *http.Response) (*http.Response, error) {
	ct := resp.Header.Get("Content-Type")

	if strings.Contains(ct, "text/event-stream") {
		pr, pw := io.Pipe()
		go func() {
			defer resp.Body.Close()
			defer pw.Close()
			_ = AnthropicStreamToOpenAI(resp.Body, pw)
		}()
		hdr := resp.Header.Clone()
		hdr.Set("Content-Type", "text/event-stream")
		hdr.Del("Content-Length")
		return wrapResponse(resp, hdr, pr), nil
	}

	raw, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	oai, err := AnthropicRespToOpenAI(raw)
	if err != nil {
		return nil, err
	}
	hdr := resp.Header.Clone()
	hdr.Set("Content-Type", "application/json")
	hdr.Del("Content-Length")
	return wrapResponse(resp, hdr, io.NopCloser(bytes.NewReader(oai))), nil
}

func wrapResponse(orig *http.Response, hdr http.Header, body io.ReadCloser) *http.Response {
	return &http.Response{
		StatusCode: orig.StatusCode,
		Status:     orig.Status,
		Header:     hdr,
		Body:       body,
		Request:    orig.Request,
	}
}
