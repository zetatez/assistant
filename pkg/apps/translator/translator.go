package translator

import (
	"context"
	"encoding/json"
	"fmt"

	"assistant/pkg/llm"
)

type Translator struct {
	client llm.Client
}

func NewTranslator(client llm.Client) *Translator {
	return &Translator{client: client}
}

func (t *Translator) Translate(
	ctx context.Context,
	text string,
	targetLang string,
) (*Result, error) {
	prompt := fmt.Sprintf(PromptTpl, targetLang, targetLang, text)
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

	content, err := llm.ExtractAndValidateJSONObject(resp.Content)
	if err != nil {
		return nil, err
	}

	var result Result
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w\nraw: %s", err, resp.Content)
	}

	return &result, nil
}

const PromptTpl = `
你是一个严格的翻译引擎，不是聊天助手，也不是解释工具。

你的唯一职责是翻译。

任务规范：
1. 自动识别输入文本的源语言
2. 判断输入文本类型，仅限以下三种之一：
   - word（单个词或短语）
   - sentence（单句或多句但不构成完整文章）
   - article（段落或文章）
3. 将输入文本翻译为目标语言：%s
4. 翻译风格要求：
   - 自然
   - 简洁
   - 书面
   - 优雅
   - 不添加任何解释、注释或说明

输出规范（必须严格遵守）：
- 仅输出一个 JSON 对象
- 不允许输出除 JSON 以外的任何字符（包括多余换行）
- 不允许使用 Markdown
- JSON 必须是合法且可直接解析的

JSON 字段定义：
{
  "source_language": "ISO 语言名称或通用语言名（如 English, Chinese, Japanese）",
  "target_language": "%s",
  "input_type": "word | sentence | article",
  "translation": "翻译后的完整文本",
  "confidence": 0.0 到 1.0 之间的小数
}

重要约束：
- 如果无法识别源语言，使用 "unknown"
- confidence 表示对翻译准确性的主观置信度
- 不要省略任何字段
- 不要更改 JSON 字段名
- 不要对输入文本进行总结或改写

待翻译文本：
%s
`
