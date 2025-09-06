package diagnoser

type Issue struct {
	Type      string `json:"type"`       // 问题类型
	Severity  string `json:"severity"`   // critical | high | medium | low
	Message   string `json:"message"`    // 问题描述
	Location  string `json:"location"`   // 位置信息
	ErrorCode string `json:"error_code"` // 错误码
	Timestamp string `json:"timestamp"`  // 时间戳
}

type RootCause struct {
	Primary             string   `json:"primary"`              // 主要根因
	Category            string   `json:"category"`             // 根因分类
	ContributingFactors []string `json:"contributing_factors"` // 次要原因
	Confidence          string   `json:"confidence"`           // high | medium | low
}

type Solution struct {
	Description     string   `json:"description"`      // 解决方案描述
	Priority        string   `json:"priority"`         // critical | high | medium | low
	Category        string   `json:"category"`         // immediate | temporary | permanent
	Actionable      bool     `json:"actionable"`       // 是否可执行
	EstimatedEffort string   `json:"estimated_effort"` // low | medium | high
	SideEffects     []string `json:"side_effects"`     // 可能的副作用
}

type Result struct {
	ProblemDomain      string     `json:"problem_domain"`      // 问题域
	ProblemType        string     `json:"problem_type"`        // 问题类型
	Severity           string     `json:"severity"`            // critical | high | medium | low
	ImpactScope        string     `json:"impact_scope"`        // single_component | multiple_components | entire_service | entire_system
	Summary            string     `json:"summary"`             // 问题简要描述
	Issues             []Issue    `json:"issues"`              // 识别到的问题列表
	RootCause          RootCause  `json:"root_cause"`          // 根因分析
	DiagnosisSteps     []string   `json:"diagnosis_steps"`     // 诊断步骤
	Solutions          []Solution `json:"solutions"`           // 解决方案建议
	AffectedComponents []string   `json:"affected_components"` // 受影响的组件
	Dependencies       []string   `json:"dependencies"`        // 相关依赖
	PreventionMeasures []string   `json:"prevention_measures"` // 预防措施
	Confidence         float64    `json:"confidence"`          // 0.0 ~ 1.0
}
