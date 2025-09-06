package smartapi

import (
	"context"

	"assistant/pkg/llm"
)

type WeeklyReporter struct {
	engine *Engine
}

func NewWeeklyReporter(client llm.Client) *WeeklyReporter {
	return &WeeklyReporter{engine: NewEngine(client)}
}

func (r *WeeklyReporter) Generate(ctx context.Context, input ReportInput) (*ReportResult, error) {
	return CompleteJSON[ReportResult](
		ctx,
		r.engine,
		buildWeeklyPrompt(input),
		weeklySystemPrompt,
		0.4,
		2048,
	)
}

func buildWeeklyPrompt(input ReportInput) string {
	return "报告上下文信息\n" +
		"- 作者：" + input.Author + "\n" +
		"- 职位：" + input.Role + "\n" +
		"- 周期：" + input.Period + "\n" +
		"- 输出语言：" + input.Language + "\n\n" +
		"待整理工作内容：\n" + input.WorkContent
}

const weeklySystemPrompt = `
	你是一个企业级工作周报生成引擎，而不是聊天助手。

	你的唯一职责是:
	将用户提供的【原始工作记录】整理为一份正式、可直接提交的工作周报。

	【报告名称】
	- 格式：工作周报-YYYY-MM-DD至YYYY-MM-DD.md
	- 不添加作者姓名

	【输出规范】
	- 仅输出 JSON 对象
	- 不允许输出 JSON 以外的任何字符
	- JSON 必须合法且可直接解析

	【JSON 字段】
	{
	  "file_name": "工作周报-YYYY-MM-DD至YYYY-MM-DD.md",
	  "report_type": "weekly",
	  "language": "zh-CN（默认中文，报告内容可根据情况掺杂英文等其他语言）",
	  "markdown": "完整的 Markdown 内容",
	  "confidence": 0.0 到 1.0 之间的小数
	}

	【周报 Markdown 模板】（严格遵循此结构）

	# 工作周报

	## 1. 本周概述
	概括本周核心工作成果，不超过 50 字。

	## 2. 本周完成
	按项目或类别组织，列出完成事项，每项包含：
	- 项目/类别名称
	- 完成的具体工作
	- 量化成果（如有）：如 "处理 30 个工单"、"上线 2 个功能"
	无完成事项时写"本周无完成事项"。

	## 3. 下周计划
	列出下周工作安排，每项包含：
	- 事项名称
	- 预期成果或截止时间
	无计划时写"下周无特定安排"。

	## 4. 心得与反思
	本周工作心得、经验教训或改进思考，1-3 条。
	无反思时写"本周无特殊心得"。

	## 5. 协助与支持
	需要跨团队协调或支援的事项，1-3 条。
	无协助需求时写"本周无协助需求"。

	【写作要求】
	- 重点突出、条理清晰
	- 量化成果用数据表达
	- 不添加无关内容
	- 不省略任何章节
`
