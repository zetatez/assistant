package smartapi

import (
	"context"
	"fmt"

	"assistant/pkg/llm"
)

type LongTextSummarizer struct {
	engine *Engine
}

func NewLongTextSummarizer(client llm.Client) *LongTextSummarizer {
	return &LongTextSummarizer{engine: NewEngine(client)}
}

type SummarizeInput struct {
	Text       string   `json:"text"`
	Style      string   `json:"style,omitempty"`
	MaxLength  int      `json:"max_length,omitempty"`
	FocusAreas []string `json:"focus_areas,omitempty"`
}

type SummarizeResult struct {
	Summary    string   `json:"summary"`
	KeyPoints  []string `json:"key_points"`
	Language   string   `json:"language"`
	Confidence float64  `json:"confidence"`
}

const longTextSummarizePrompt = `
	【摘要风格说明】
	- brief：简洁摘要，一段话概括核心
	- standard：标准摘要，3-5 个要点
	- detailed：详细摘要，全面覆盖各部分
	- bullet：要点列表形式

	【输出规范】
	- 仅输出 JSON 对象
	- 不允许输出 JSON 以外的任何字符
	- JSON 必须合法且可直接解析

	【JSON 字段】
	{
	  "summary": "核心摘要内容",
	  "key_points": ["要点1", "要点2", ...],
	  "language": "zh-CN（默认中文，报告内容可根据情况掺杂英文等其他语言）",
	  "confidence": 0.0 到 1.0 之间的小数
	}

	【写作要求】
	- 语言简洁、准确
	- key_points 数量控制在 3-8 个
	- 保留原文关键信息和数据
	- 不添加无关内容
	- 不省略任何字段
`

func (s *LongTextSummarizer) Summarize(ctx context.Context, input SummarizeInput) (*SummarizeResult, error) {
	style := input.Style
	if style == "" {
		style = "standard"
	}

	maxLen := input.MaxLength
	if maxLen == 0 {
		maxLen = 500
	}

	prompt := s.buildSummarizePrompt(input)
	systemPrompt := s.buildSystemPrompt(style, maxLen)

	return CompleteJSON[SummarizeResult](
		ctx,
		s.engine,
		prompt,
		systemPrompt,
		0.3,
		2048,
	)
}

func (s *LongTextSummarizer) buildSystemPrompt(style string, maxLength int) string {
	return fmt.Sprintf(longTextSummarizePrompt+`

	【本次任务要求】
	- 摘要风格：%s
	- 摘要最大长度：%d 字
	`, style, maxLength)
}

func (s *LongTextSummarizer) buildSummarizePrompt(input SummarizeInput) string {
	prompt := "待摘要文本：\n" + input.Text

	if len(input.FocusAreas) > 0 {
		prompt += "\n\n重点关注领域：\n"
		for _, area := range input.FocusAreas {
			prompt += "- " + area + "\n"
		}
	}

	return prompt
}
