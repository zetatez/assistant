package reporter

type ReportType string

const (
	WeeklyReport  ReportType = "weekly"
	MonthlyReport ReportType = "monthly"
	QuarterReport ReportType = "quarterly"
	YearlyReport  ReportType = "yearly"
)

type Input struct {
	ReportType ReportType `json:"report_type"`

	// 基础信息（用于上下文与表述，不用于虚构内容）
	Author   string `json:"author"`
	Role     string `json:"role"`
	Period   string `json:"period"` // 例如：2026-01-01 ~ 2026-01-07
	Language string `json:"language"`

	// 唯一事实来源
	WorkContent string `json:"work_content"`
}

type Result struct {
	FileName   string  `json:"file_name"` // 报告名称
	ReportType string  `json:"report_type"`
	Language   string  `json:"language"`
	Markdown   string  `json:"markdown"`
	Confidence float64 `json:"confidence"`
}
