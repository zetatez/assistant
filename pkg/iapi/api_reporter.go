package iapi

import (
	"context"
	"fmt"

	"assistant/pkg/llm"
)

type Reporter struct {
	engine *LLMEngine
}

func NewReporter(client llm.Client) *Reporter {
	return &Reporter{engine: NewLLMEngine(client)}
}

func (r *Reporter) Generate(ctx context.Context, input ReportInput) (*ReportResult, error) {
	prompt := fmt.Sprintf(
		reportPromptTpl,
		input.ReportType,
		input.Author,
		input.Role,
		input.Period,
		input.Language,
		input.WorkContent,
	)

	return CompleteJSON[ReportResult](
		ctx,
		r.engine,
		prompt,
		reportSystemPrompt,
		0.4,
		2048,
	)
}

type ReportType string

const (
	WeeklyReport  ReportType = "weekly"
	MonthlyReport ReportType = "monthly"
	QuarterReport ReportType = "quarterly"
	YearlyReport  ReportType = "yearly"
)

type ReportInput struct {
	ReportType ReportType `json:"report_type"`

	// 基础信息（用于上下文与表述，不用于虚构内容）
	Author   string `json:"author"`
	Role     string `json:"role"`
	Period   string `json:"period"` // 例如：2026-01-01 ~ 2026-01-07
	Language string `json:"language"`

	// 唯一事实来源
	WorkContent string `json:"work_content"`
}

type ReportResult struct {
	FileName   string  `json:"file_name"` // 报告名称
	ReportType string  `json:"report_type"`
	Language   string  `json:"language"`
	Markdown   string  `json:"markdown"`
	Confidence float64 `json:"confidence"`
}

const reportSystemPrompt = `
	你是一个企业级工作报告生成引擎，而不是聊天助手。

	你的唯一职责是:
	将用户提供的【原始工作记录】整理为一份正式、可直接提交的工作报告。

	强制约束（必须遵守）
	- 报告类型由系统指定，不允许自行更改
	- 报告类型只能是以下之一:
	  - weekly（工作周报）
	  - monthly（工作月报）
	  - quarterly（工作季报）
	  - yearly（年终总结）

	报告名称生成规则（必须遵守）
	- 必须生成一个 文件名称
	- 文件名称格式：
	  {报告类型中文名}-{开始时间}.{结束时间}.md
	- 报告类型中文映射如下：
	  - weekly    → 工作周报
	  - monthly   → 工作月报
	  - quarterly → 工作季报
	  - yearly    → 年终工作总结
	- 不要添加作者姓名到名称中
	- 名称必须简洁、正式、可用于文件名

	输出格式规范（必须严格遵守）
	- 仅输出一个 JSON 对象
	- 不允许输出除 JSON 以外的任何字符（包括多余换行）
	- 不允许使用 Markdown
	- JSON 必须是合法且可直接解析的

	JSON 字段定义:
	{
	  "file_name": "文件名称",
	  "report_type": "weekly | monthly | quarterly | yearly",
	  "language": "输出语言，如 zh-CN 或 en-US",
	  "markdown": "完整的 Markdown 报告内容",
	  "confidence": 0.0 到 1.0 之间的小数
	}

	重要约束:
	- 用户提供的原始工作记录（唯一信息来源）
	- 不要省略任何字段
	- 不要更改 JSON 字段名
`

const reportPromptTpl = `
	报告上下文信息
	- 报告类型：%s
	- 作者：%s
	- 职位：%s
	- 报告周期：%s
	- 输出语言：%s

	待整理工作内容：
	%s
`
