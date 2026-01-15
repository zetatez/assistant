package llmtranslator

import (
	"context"
	"encoding/json"
	"fmt"

	"assistant/pkg/llm"
)

type Result struct {
	SourceLanguage string  `json:"source_language"`
	TargetLanguage string  `json:"target_language"`
	InputType      string  `json:"input_type"`
	Translation    string  `json:"translation"`
	Confidence     float64 `json:"confidence"`
}

type Translator struct {
	client llm.Client
}

func New(client llm.Client) *Translator {
	return &Translator{client: client}
}

func (t *Translator) Translate(
	ctx context.Context,
	text string,
	targetLang string,
) (*Result, error) {

	prompt := fmt.Sprintf(`
你是一个专业翻译引擎，而不是聊天助手。

任务：
1. 自动识别输入文本的语言
2. 判断输入是：单词、句子或文章
3. 将其翻译为 %s
4. 翻译要求：自然、简洁、书面、优雅、无多余解释

输出要求：
- 只返回 JSON
- 不要包含任何额外文字
- 不要使用 Markdown
- JSON 必须可被直接解析

JSON 格式：
{
  "source_language": "...",
  "target_language": "%s",
  "input_type": "word | sentence | article",
  "translation": "...",
  "confidence": 0.0-1.0
}

待翻译文本：
%s
`, targetLang, targetLang, text)

	req := llm.ChatRequest{
		Model: t.client.Model(),
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: prompt},
		},
		Temperature: 0.2, // 翻译场景要低
		MaxTokens:   512,
	}

	resp, err := t.client.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	var result Result
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w\nraw: %s", err, resp.Content)
	}

	return &result, nil
}
