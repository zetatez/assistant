package sqladvisor

type Result struct {
	DatabaseType  string   `json:"database_type"` // oceanbase_mysql / mysql / postgres / sqlite / unknown
	OriginalSQL   string   `json:"original_sql"`
	OptimizedSQL  string   `json:"optimized_sql"`
	Optimizations []string `json:"optimizations"` // 具体做了哪些优化
	RiskLevel     string   `json:"risk_level"`    // low / medium / high
	Confidence    float64  `json:"confidence"`    // 0.0 ~ 1.0
}
