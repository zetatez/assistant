package sqlindexadvisor

type Input struct {
	SQL          string   `json:"sql"`
	TableDDLs    []string `json:"table_ddls"`
	StatsText    string   `json:"stats_text,omitempty"` // NDV / row count / explain 等说明
	ExtraContext string   `json:"extra_context,omitempty"`
}

type IndexDDL struct {
	DDL    string `json:"ddl"`
	Reason string `json:"reason"`
	Risk   string `json:"risk"` // low / medium / high
}

type TableIndexPlan struct {
	TableName string     `json:"table_name"`
	Actions   []IndexDDL `json:"actions"` // CREATE / DROP / ALTER INDEX
}

type Result struct {
	DatabaseType string           `json:"database_type"`
	OriginalSQL  string           `json:"original_sql"`
	Plans        []TableIndexPlan `json:"plans"`
	GlobalRisk   string           `json:"global_risk"`
	Confidence   float64          `json:"confidence"`
}
