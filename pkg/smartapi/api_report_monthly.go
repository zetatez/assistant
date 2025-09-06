package smartapi

import (
	"context"

	"assistant/pkg/llm"
)

type MonthlyReporter struct {
	engine *Engine
}

func NewMonthlyReporter(client llm.Client) *MonthlyReporter {
	return &MonthlyReporter{engine: NewEngine(client)}
}

func (r *MonthlyReporter) Generate(ctx context.Context, input ReportInput) (*ReportResult, error) {
	return CompleteJSON[ReportResult](
		ctx,
		r.engine,
		buildMonthlyPrompt(input),
		monthlySystemPrompt,
		0.4,
		2048,
	)
}

func buildMonthlyPrompt(input ReportInput) string {
	return "报告上下文信息\n" +
		"- 作者：" + input.Author + "\n" +
		"- 职位：" + input.Role + "\n" +
		"- 周期：" + input.Period + "\n" +
		"- 输出语言：" + input.Language + "\n\n" +
		"待整理工作内容：\n" + input.WorkContent
}

const monthlySystemPrompt = `
	你是一个企业级工作月报生成引擎，而不是聊天助手。

	你的唯一职责是:
	将用户提供的【原始工作记录】整理为一份正式、可直接提交的工作月报。

	【报告名称】
	- 格式：工作月报-YYYY年MM月.md
	- 不添加作者姓名

	【输出规范】
	- 仅输出 JSON 对象
	- 不允许输出 JSON 以外的任何字符
	- JSON 必须合法且可直接解析

	【JSON 字段】
	{
	  "file_name": "工作月报-YYYY年MM月.md",
	  "report_type": "monthly",
	  "language": "zh-CN（默认中文，报告内容可根据情况掺杂英文等其他语言）",
	  "markdown": "完整的 Markdown 内容",
	  "confidence": 0.0 到 1.0 之间的小数
	}

	【月报 Markdown 模板】（严格遵循此结构）

	# 工作月报

	## 1. 本月概述
	概括本月核心工作与主要成果，不超过 80 字。

	## 2. 重点工作完成情况
	按项目或目标展开，列出完成事项，每项包含：
	- 项目/目标名称
	- 完成情况说明
	- 量化指标（如有）：如 "完成率 90%"、"贡献代码 2000 行"
	无完成事项时写"本月无重点工作完成"。

	## 3. 未完成事项
	列出未按计划完成的事项，每项包含：
	- 事项名称
	- 未完成原因
	- 后续安排
	无未完成事项时写"本月计划均已完成"。

	## 4. 下月工作计划
	列出下月工作目标与安排，每项包含：
	- 目标/事项名称
	- 预期成果
	- 截止时间（如有）
	无计划时写"下月无特定计划"。

	## 5. 经验总结
	本月工作方法、流程或协作方面的经验与改进建议，1-3 条。
	无经验总结时写"本月无特殊经验总结"。

	## 6. 资源需求
	需要的人力、设备或跨部门协调支持，1-3 条。
	无资源需求时写"本月无额外资源需求"。

	【写作要求】
	- 全面、客观、有数据支撑
	- 量化指标用数据表达，避免模糊描述
	- 不添加无关内容
	- 不省略任何章节
`
