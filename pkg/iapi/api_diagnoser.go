package iapi

import (
	"context"
	"fmt"

	"assistant/pkg/llm"
)

type Diagnoser struct {
	engine *LLMEngine
}

func NewDiagnoser(client llm.Client) *Diagnoser {
	return &Diagnoser{engine: NewLLMEngine(client)}
}

func (d *Diagnoser) Diagnose(
	ctx context.Context,
	input string,
) (*DiagnoseResult, error) {
	prompt := fmt.Sprintf(diagnosePromptTpl, input)
	return CompleteJSON[DiagnoseResult](
		ctx,
		d.engine,
		prompt,
		diagnoseSystemPrompt,
		0.2,
		3000,
	)
}

type DiagnoseIssue struct {
	Type      string `json:"type"`       // 问题类型
	Severity  string `json:"severity"`   // critical | high | medium | low
	Message   string `json:"message"`    // 问题描述
	Location  string `json:"location"`   // 位置信息
	ErrorCode string `json:"error_code"` // 错误码
	Timestamp string `json:"timestamp"`  // 时间戳
}

type DiagnoseRootCause struct {
	Primary             string   `json:"primary"`              // 主要根因
	Category            string   `json:"category"`             // 根因分类
	ContributingFactors []string `json:"contributing_factors"` // 次要原因
	Confidence          string   `json:"confidence"`           // high | medium | low
}

type DiagnoseSolution struct {
	Description     string   `json:"description"`      // 解决方案描述
	Priority        string   `json:"priority"`         // critical | high | medium | low
	Category        string   `json:"category"`         // immediate | temporary | permanent
	Actionable      bool     `json:"actionable"`       // 是否可执行
	EstimatedEffort string   `json:"estimated_effort"` // low | medium | high
	SideEffects     []string `json:"side_effects"`     // 可能的副作用
}

type DiagnoseResult struct {
	ProblemDomain      string             `json:"problem_domain"`      // 问题域
	ProblemType        string             `json:"problem_type"`        // 问题类型
	Severity           string             `json:"severity"`            // critical | high | medium | low
	ImpactScope        string             `json:"impact_scope"`        // single_component | multiple_components | entire_service | entire_system
	Summary            string             `json:"summary"`             // 问题简要描述
	Issues             []DiagnoseIssue    `json:"issues"`              // 识别到的问题列表
	RootCause          DiagnoseRootCause  `json:"root_cause"`          // 根因分析
	DiagnosisSteps     []string           `json:"diagnosis_steps"`     // 诊断步骤
	Solutions          []DiagnoseSolution `json:"solutions"`           // 解决方案建议
	AffectedComponents []string           `json:"affected_components"` // 受影响的组件
	Dependencies       []string           `json:"dependencies"`        // 相关依赖
	PreventionMeasures []string           `json:"prevention_measures"` // 预防措施
	Confidence         float64            `json:"confidence"`          // 0.0 ~ 1.0
}

const diagnoseSystemPrompt = `
	你是一个专业的问题诊断引擎。分析以下问题，返回详细的 JSON 结果。

	任务要求：
	1. 识别问题域和类型
	2. 提取关键错误信息、错误码、堆栈跟踪
	3. 分析根因和影响范围
	4. 提供可执行的诊断步骤和解决方案

	问题域：hardware/software/network/data/application/system/configuration/code/security/infrastructure/cloud/mixed/unknown

	问题类型分类：
	硬件：disk_failure, memory_failure, cpu_overheat, power_failure, storage_exhaustion, io_bottleneck
	数据库：database_connection, database_deadlock, database_slow_query, database_replication, database_corruption
	应用：application_crash, out_of_memory, memory_leak, thread_deadlock, cpu_spike
	网络：network_connectivity, network_latency, dns_resolution, firewall_block, ssl_certificate
	代码：null_pointer, race_condition, deadlock, logic_error, buffer_overflow
	系统：kernel_panic, service_down, resource_exhaustion, zombie_process
	配置：misconfiguration, permission_denied, certificate_expired
	安全：authentication_failure, sql_injection, ddos_attack, data_breach
	容器：container_crash, pod_crash, container_oom, resource_quota_exceeded
	性能：high_response_time, low_throughput, memory_pressure

	根因分类：hardware_failure, software_bug, misconfiguration, resource_limitation, network_issue, human_error, external_dependency

	仅输出 JSON（无其他文本）：
	{
	  "problem_domain": "问题域",
	  "problem_type": "问题类型",
	  "severity": "critical|high|medium|low",
	  "impact_scope": "single_component|multiple_components|entire_service|entire_system",
	  "summary": "问题简要描述",
	  "issues": [
	    {"type": "问题类型", "severity": "严重级别", "message": "问题描述", "location": "位置", "error_code": "错误码", "timestamp": "时间戳"}
	  ],
	  "root_cause": {
	    "primary": "主要根因",
	    "category": "根因分类",
	    "contributing_factors": ["次要原因"],
	    "confidence": "high|medium|low"
	  },
	  "diagnosis_steps": ["诊断步骤1", "诊断步骤2"],
	  "solutions": [
	    {"description": "解决方案描述", "priority": "critical|high|medium|low", "category": "immediate|temporary|permanent", "actionable": true, "estimated_effort": "low|medium|high", "side_effects": ["副作用"]}
	  ],
	  "affected_components": ["组件1", "组件2"],
	  "dependencies": ["依赖项"],
	  "prevention_measures": ["预防措施"],
	  "confidence": 0.0-1.0
	}
`

const diagnosePromptTpl = `
	待诊断信息：
	%s
`
