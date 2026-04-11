package wiki

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"assistant/pkg/llm"
)

type Reranker interface {
	Rerank(ctx context.Context, query string, results []*GrepHit) ([]*RerankResult, error)
}

type RerankResult struct {
	Entry   *IndexedEntry
	Snippet string
	Score   float64
	Reason  string
}

type LLMReranker struct {
	client llm.Client
	model  string
}

func NewLLMReranker(client llm.Client, model string) Reranker {
	return &LLMReranker{client: client, model: model}
}

func (r *LLMReranker) Rerank(ctx context.Context, query string, results []*GrepHit) ([]*RerankResult, error) {
	if r.client == nil || len(results) == 0 {
		out := make([]*RerankResult, len(results))
		for i, res := range results {
			out[i] = &RerankResult{Entry: res.Entry, Snippet: res.Snippet, Score: float64(res.Score), Reason: ""}
		}
		return out, nil
	}

	queryEscaped := escapeForPrompt(query)

	var sb strings.Builder
	sb.WriteString("请根据以下候选文档片断，判断每个文档与查询的相关程度，并给出相关性评分(0-10分)和简要理由。\n")
	sb.WriteString("评分标准：0分=完全不相关，5分=部分相关，10分=高度相关。\n")
	sb.WriteString("如果文档与查询无关或矛盾，请给0分并说明原因。\n")
	sb.WriteString("查询: " + queryEscaped + "\n\n")

	for i, res := range results {
		sb.WriteString(strings.Repeat("=", 60) + "\n")
		sb.WriteString(fmt.Sprintf("候选文档%d:\n", i+1))
		sb.WriteString("标题: " + res.Entry.Title + "\n")
		sb.WriteString("路径: " + res.Entry.Path + "\n")
		sb.WriteString("内容片段: " + res.Snippet + "\n\n")
	}

	prompt := sb.String() + strings.Repeat("=", 60) + "\n请以JSON数组格式输出评分结果，格式如下：\n[{\"index\":0,\"score\":8.5,\"reason\":\"文档讨论了相关概念\"},...]\n直接输出JSON，不要其他内容。只输出与查询真正相关的文档，不相关的请给0分。"

	messages := []llm.Message{
		{Role: llm.RoleUser, Content: prompt},
	}

	req := llm.ChatRequest{
		Model:       r.model,
		Messages:    messages,
		Temperature: 0.1,
		MaxTokens:   1000,
	}

	resp, err := r.client.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	return parseRerankResults(resp.Content, results)
}

func escapeForPrompt(s string) string {
	js, err := json.Marshal(s)
	if err != nil {
		return s
	}
	return string(js)
}

func parseRerankResults(jsonStr string, results []*GrepHit) ([]*RerankResult, error) {
	jsonStr = strings.TrimSpace(jsonStr)

	start := strings.Index(jsonStr, "[")
	end := strings.LastIndex(jsonStr, "]")
	if start == -1 || end == -1 || start >= end {
		out := make([]*RerankResult, len(results))
		for i, res := range results {
			out[i] = &RerankResult{Entry: res.Entry, Snippet: res.Snippet, Score: float64(res.Score), Reason: ""}
		}
		return out, nil
	}

	arrContent := jsonStr[start+1 : end]
	lines := strings.Split(arrContent, ",")

	var reranked []*RerankResult
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue
		}

		idx := parseJSONInt(line, "index")
		score := parseJSONFloat(line, "score")
		reason := parseJSONString(line, "reason")

		if idx >= 0 && idx < len(results) {
			reranked = append(reranked, &RerankResult{
				Entry:   results[idx].Entry,
				Snippet: results[idx].Snippet,
				Score:   score,
				Reason:  reason,
			})
		}
	}

	if len(reranked) == 0 {
		reranked = make([]*RerankResult, len(results))
		for i, res := range results {
			reranked[i] = &RerankResult{Entry: res.Entry, Snippet: res.Snippet, Score: float64(res.Score), Reason: ""}
		}
	}

	return reranked, nil
}

func parseJSONInt(line, key string) int {
	prefix := `"` + key + `":`
	idx := strings.Index(line, prefix)
	if idx == -1 {
		return -1
	}
	rest := strings.TrimSpace(line[idx+len(prefix):])
	end := len(rest)
	for i, c := range rest {
		if c < '0' || c > '9' {
			end = i
			break
		}
	}
	if end == 0 {
		return -1
	}
	var n int
	for _, c := range rest[:end] {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

func parseJSONFloat(line, key string) float64 {
	prefix := `"` + key + `":`
	idx := strings.Index(line, prefix)
	if idx == -1 {
		return 0
	}
	rest := strings.TrimSpace(line[idx+len(prefix):])
	var numStr string
	for _, c := range rest {
		if (c >= '0' && c <= '9') || c == '.' {
			numStr += string(c)
		} else {
			break
		}
	}
	if numStr == "" {
		return 0
	}
	var f float64
	for _, c := range numStr {
		if c == '.' {
			break
		}
		f = f*10 + float64(c-'0')
	}
	if dp := strings.Index(numStr, "."); dp != -1 && dp < len(numStr)-1 {
		fracStr := numStr[dp+1:]
		frac := 0.0
		div := 1.0
		for _, c := range fracStr {
			if c >= '0' && c <= '9' {
				frac = frac*10 + float64(c-'0')
				div *= 10
			}
		}
		f += frac / div
	}
	return f
}

func parseJSONString(line, key string) string {
	prefix := `"` + key + `":"`
	idx := strings.Index(line, prefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(prefix)
	end := start
	for end < len(line) {
		if line[end] == '"' && (end+1 >= len(line) || line[end+1] == ',' || line[end+1] == '}') {
			break
		}
		end++
	}
	return strings.TrimSpace(line[start:end])
}
