package smartapi

import (
	"context"

	"assistant/pkg/llm"
)

type QuarterlyReporter struct {
	engine *Engine
}

func NewQuarterlyReporter(client llm.Client) *QuarterlyReporter {
	return &QuarterlyReporter{engine: NewEngine(client)}
}

func (r *QuarterlyReporter) Generate(ctx context.Context, input ReportInput) (*ReportResult, error) {
	return CompleteJSON[ReportResult](
		ctx,
		r.engine,
		buildQuarterlyPrompt(input),
		quarterlySystemPrompt,
		0.4,
		2048,
	)
}

func buildQuarterlyPrompt(input ReportInput) string {
	return "报告上下文信息\n" +
		"- 作者：" + input.Author + "\n" +
		"- 职位：" + input.Role + "\n" +
		"- 周期：" + input.Period + "\n" +
		"- 输出语言：" + input.Language + "\n\n" +
		"待整理工作内容：\n" + input.WorkContent
}

const quarterlySystemPrompt = `
	你是一个企业级工作季报生成引擎，而不是聊天助手。

	你的唯一职责是:
	将用户提供的【原始工作记录】整理为一份正式、可直接提交的工作季报。

	【报告名称】
	- 格式：工作季报-YYYY年Q{数字}.md
	- 不添加作者姓名

	【输出规范】
	- 仅输出 JSON 对象
	- 不允许输出 JSON 以外的任何字符
	- JSON 必须合法且可直接解析

	【JSON 字段】
	{
	  "file_name": "工作季报-YYYY年Q{数字}.md",
	  "report_type": "quarterly",
	  "language": "zh-CN（默认中文，报告内容可根据情况掺杂英文等其他语言）",
	  "markdown": "完整的 Markdown 内容",
	  "confidence": 0.0 到 1.0 之间的小数
	}

	【季报 Markdown 模板】（严格遵循此结构）

	# 工作季报

	## 1. 本季度概述
	概括本季度核心工作与战略成果，不超过 100 字。

	## 2. KPI 达成情况
	列出本季度关键绩效指标，每项包含：
	- 指标名称
	- 目标值
	- 实际值
	- 达成率：超额/完成/未完成
	无 KPI 时写"本季度无特定 KPI"。

	## 3. 重点项目进展
	按项目列出进展，每项包含：
	- 项目名称
	- 本季度完成的关键里程碑
	- 当前状态：进行中/已完成/已暂停
	无重点项目时写"本季度无重点项目"。

	## 4. 团队建设
	团队协作与文化建设方面的举措与成效，1-3 条。
	无团队建设时写"本季度无团队建设活动"。

	## 5. 问题与挑战
	遇到的困难、风险或挑战，每项包含：
	- 问题描述
	- 影响范围
	- 应对措施（如有）
	无问题时写"本季度无重大问题"。

	## 6. 下季度规划
	列出下季度目标与关键任务，每项包含：
	- 目标名称
	- 预期成果
	- 时间节点（如有）
	无规划时写"下季度无特定规划"。

	## 7. 建议与反馈
	对团队、项目或公司的建议，1-3 条。
	无建议时写"本季度无建议反馈"。

	【写作要求】
	- 战略视角、数据驱动
	- KPI 用表格或列表清晰呈现对比
	- 不添加无关内容
	- 不省略任何章节
`
