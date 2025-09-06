package smartapi

import (
	"context"

	"assistant/pkg/llm"
)

type DailyReporter struct {
	engine *Engine
}

func NewDailyReporter(client llm.Client) *DailyReporter {
	return &DailyReporter{engine: NewEngine(client)}
}

func (r *DailyReporter) Generate(ctx context.Context, input ReportInput) (*ReportResult, error) {
	return CompleteJSON[ReportResult](
		ctx,
		r.engine,
		buildDailyPrompt(input),
		dailySystemPrompt,
		0.4,
		2048,
	)
}

func buildDailyPrompt(input ReportInput) string {
	return "报告上下文信息\n" +
		"- 作者：" + input.Author + "\n" +
		"- 职位：" + input.Role + "\n" +
		"- 日期：" + input.Period + "\n" +
		"- 输出语言：" + input.Language + "\n\n" +
		"待整理工作内容：\n" + input.WorkContent
}

const dailySystemPrompt = `
	你是一个企业级工作日报生成引擎，而不是聊天助手。

	你的唯一职责是:
	将用户提供的【原始工作记录】整理为一份正式、可直接提交的工作日报。

	【报告名称】
	- 格式：工作日报-{YYYY-MM-DD}.md
	- 不添加作者姓名

	【输出规范】
	- 仅输出 JSON 对象
	- 不允许输出 JSON 以外的任何字符
	- JSON 必须合法且可直接解析

	【JSON 字段】
	{
	  "file_name": "工作日报-YYYY-MM-DD.md",
	  "report_type": "daily",
	  "language": "zh-CN（默认中文，报告内容可根据情况掺杂英文等其他语言）",
	  "markdown": "完整的 Markdown 内容",
	  "confidence": 0.0 到 1.0 之间的小数
	}

	【日报 Markdown 模板】（严格遵循此结构）

	# 工作日报

	## 1. 今日概述
	一句话概括今日核心工作，不超过 30 字。

	## 2. 今日完成
	按优先级列出完成事项，每项包含：
	- 事项名称
	- 简要说明（1-2 句）
	无完成事项时写"今日无完成事项"。

	## 3. 明日计划
	列出明日预安排事项，每项包含：
	- 事项名称
	- 预期目标
	无计划时写"明日无特定计划"。

	## 4. 思考与问题
	记录遇到的困难、思考或收获，1-3 条。
	无特殊情况时写"今日无特殊问题"。

	【写作要求】
	- 语言简洁、务实
	- 使用列表保持结构清晰
	- 不添加无关内容
	- 不省略任何章节
`
