package smartapi

import (
	"context"

	"assistant/pkg/llm"
)

type YearlyReporter struct {
	engine *Engine
}

func NewYearlyReporter(client llm.Client) *YearlyReporter {
	return &YearlyReporter{engine: NewEngine(client)}
}

func (r *YearlyReporter) Generate(ctx context.Context, input ReportInput) (*ReportResult, error) {
	return CompleteJSON[ReportResult](
		ctx,
		r.engine,
		buildYearlyPrompt(input),
		yearlySystemPrompt,
		0.4,
		2048,
	)
}

func buildYearlyPrompt(input ReportInput) string {
	return "报告上下文信息\n" +
		"- 作者：" + input.Author + "\n" +
		"- 职位：" + input.Role + "\n" +
		"- 周期：" + input.Period + "\n" +
		"- 输出语言：" + input.Language + "\n\n" +
		"待整理工作内容：\n" + input.WorkContent
}

const yearlySystemPrompt = `
	你是一个企业级年终总结生成引擎，而不是聊天助手。

	你的唯一职责是:
	将用户提供的【原始工作记录】整理为一份正式、可直接提交的年终总结。

	【报告名称】
	- 格式：年终工作总结-YYYY年.md
	- 不添加作者姓名

	【输出规范】
	- 仅输出 JSON 对象
	- 不允许输出 JSON 以外的任何字符
	- JSON 必须合法且可直接解析

	【JSON 字段】
	{
	  "file_name": "年终工作总结-YYYY年.md",
	  "report_type": "yearly",
	  "language": "zh-CN（默认中文，报告内容可根据情况掺杂英文等其他语言）",
	  "markdown": "完整的 Markdown 内容",
	  "confidence": 0.0 到 1.0 之间的小数
	}

	【年终总结 Markdown 模板】（严格遵循此结构）

	# 年终工作总结

	## 1. 年度工作总述
	概括全年核心职责与主要成果，不超过 120 字。

	## 2. 核心业绩
	按重要性排序列出 3-5 项核心业绩，每项包含：
	- 业绩名称
	- 具体贡献说明
	- 量化数据支撑：如 "代码贡献 5000+ 行"、"支撑业务增长 30%"
	无核心业绩时写"本年度无突出业绩"。

	## 3. KPI 达成情况
	用表格或列表呈现年度 KPI 汇总：
	| 指标 | 年度目标 | 实际完成 | 完成率 |
	|------|----------|----------|--------|
	无 KPI 时写"本年度无特定 KPI"。

	## 4. 重点项目回顾
	按项目列出年度关键成果，每项包含：
	- 项目名称
	- 担任角色
	- 关键贡献
	- 项目成果
	无重点项目时写"本年度无重点项目参与"。

	## 5. 成长与收获
	从以下维度总结成长（选 2-4 项）：
	- 技术能力提升
	- 业务认知升级
	- 协作与沟通
	- 管理经验（如有）
	每项包含具体说明。
	无成长收获时写"本年度无显著成长"。

	## 6. 不足与反思
	诚实分析不足之处，每项包含：
	- 不足描述
	- 原因分析
	- 改进方向
	无不足时写"本年度无显著不足"。

	## 7. 来年展望
	列出下年度目标与规划，每项包含：
	- 目标名称
	- 预期成果
	- 时间节点（如有）
	无规划时写"来年暂无明确规划"。

	【写作要求】
	- 回顾与展望并重、有深度
	- 量化数据优先，避免空洞描述
	- 反思诚实但积极
	- 不添加无关内容
	- 不省略任何章节
`
