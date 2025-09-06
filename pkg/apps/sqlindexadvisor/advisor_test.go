package sqlindexadvisor

import (
	"testing"
)

func TestNewSQLAdvisor(t *testing.T) {
	advisor := NewSQLAdvisor(nil)

	if advisor == nil {
		t.Fatal("expected non-nil SQLAdvisor even with nil client")
	}
}

func TestInput_Fields(t *testing.T) {
	input := Input{
		SQL:          "SELECT * FROM users WHERE name = ? AND status = ?",
		TableDDLs:    []string{"CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(255), status INT)"},
		StatsText:    "Table has 1000000 rows, name has high cardinality (NDV=500000)",
		ExtraContext: "Frequent queries on status = 0",
	}

	if input.SQL != "SELECT * FROM users WHERE name = ? AND status = ?" {
		t.Errorf("unexpected SQL: %s", input.SQL)
	}

	if len(input.TableDDLs) != 1 {
		t.Errorf("expected 1 table DDL, got %d", len(input.TableDDLs))
	}

	if input.StatsText != "Table has 1000000 rows, name has high cardinality (NDV=500000)" {
		t.Errorf("unexpected StatsText: %s", input.StatsText)
	}

	if input.ExtraContext != "Frequent queries on status = 0" {
		t.Errorf("unexpected ExtraContext: %s", input.ExtraContext)
	}
}

func TestInput_EmptyOptionalFields(t *testing.T) {
	input := Input{
		SQL:       "SELECT * FROM users WHERE id = ?",
		TableDDLs: []string{"CREATE TABLE users (id INT PRIMARY KEY)"},
	}

	if input.StatsText != "" {
		t.Errorf("expected empty StatsText, got '%s'", input.StatsText)
	}

	if input.ExtraContext != "" {
		t.Errorf("expected empty ExtraContext, got '%s'", input.ExtraContext)
	}
}

func TestIndexDDL_Fields(t *testing.T) {
	indexDDL := IndexDDL{
		DDL:    "CREATE INDEX idx_users_name ON users(name)",
		Reason: "High cardinality column frequently used in WHERE clause",
		Risk:   "low",
	}

	if indexDDL.DDL != "CREATE INDEX idx_users_name ON users(name)" {
		t.Errorf("unexpected DDL: %s", indexDDL.DDL)
	}

	if indexDDL.Reason != "High cardinality column frequently used in WHERE clause" {
		t.Errorf("unexpected Reason: %s", indexDDL.Reason)
	}

	if indexDDL.Risk != "low" {
		t.Errorf("expected Risk 'low', got '%s'", indexDDL.Risk)
	}
}

func TestIndexDDL_RiskLevels(t *testing.T) {
	riskLevels := []string{"low", "medium", "high"}

	for _, risk := range riskLevels {
		indexDDL := IndexDDL{
			DDL:  "CREATE INDEX test ON t(c)",
			Risk: risk,
		}

		if indexDDL.Risk != risk {
			t.Errorf("expected Risk '%s', got '%s'", risk, indexDDL.Risk)
		}
	}
}

func TestTableIndexPlan_Fields(t *testing.T) {
	plan := TableIndexPlan{
		TableName: "users",
		Actions: []IndexDDL{
			{
				DDL:    "CREATE INDEX idx_users_name ON users(name)",
				Reason: "Optimize WHERE clause on name",
				Risk:   "low",
			},
		},
	}

	if plan.TableName != "users" {
		t.Errorf("expected TableName 'users', got '%s'", plan.TableName)
	}

	if len(plan.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(plan.Actions))
	}

	if plan.Actions[0].DDL != "CREATE INDEX idx_users_name ON users(name)" {
		t.Errorf("unexpected DDL in action: %s", plan.Actions[0].DDL)
	}
}

func TestTableIndexPlan_EmptyActions(t *testing.T) {
	plan := TableIndexPlan{
		TableName: "users",
		Actions:   []IndexDDL{},
	}

	if len(plan.Actions) != 0 {
		t.Errorf("expected 0 actions, got %d", len(plan.Actions))
	}
}

func TestResult_Fields(t *testing.T) {
	result := Result{
		DatabaseType: "mysql",
		OriginalSQL:  "SELECT * FROM users WHERE name = ?",
		Plans: []TableIndexPlan{
			{
				TableName: "users",
				Actions: []IndexDDL{
					{
						DDL:    "CREATE INDEX idx_users_name ON users(name)",
						Reason: "Optimize WHERE clause",
						Risk:   "low",
					},
				},
			},
		},
		GlobalRisk: "low",
		Confidence: 0.88,
	}

	if result.DatabaseType != "mysql" {
		t.Errorf("expected DatabaseType 'mysql', got '%s'", result.DatabaseType)
	}

	if result.OriginalSQL != "SELECT * FROM users WHERE name = ?" {
		t.Errorf("unexpected OriginalSQL: %s", result.OriginalSQL)
	}

	if len(result.Plans) != 1 {
		t.Errorf("expected 1 plan, got %d", len(result.Plans))
	}

	if result.GlobalRisk != "low" {
		t.Errorf("expected GlobalRisk 'low', got '%s'", result.GlobalRisk)
	}

	if result.Confidence != 0.88 {
		t.Errorf("expected Confidence 0.88, got %f", result.Confidence)
	}
}

func TestResult_MultipleTables(t *testing.T) {
	result := Result{
		DatabaseType: "mysql",
		OriginalSQL:  "SELECT * FROM users u JOIN orders o ON u.id = o.user_id WHERE u.name = ?",
		Plans: []TableIndexPlan{
			{
				TableName: "users",
				Actions: []IndexDDL{
					{
						DDL:    "CREATE INDEX idx_users_name ON users(name)",
						Reason: "Filter by name",
						Risk:   "low",
					},
				},
			},
			{
				TableName: "orders",
				Actions: []IndexDDL{
					{
						DDL:    "CREATE INDEX idx_orders_user_id ON orders(user_id)",
						Reason: "Join condition",
						Risk:   "low",
					},
				},
			},
		},
		GlobalRisk: "low",
		Confidence: 0.92,
	}

	if len(result.Plans) != 2 {
		t.Errorf("expected 2 plans, got %d", len(result.Plans))
	}

	if result.Plans[0].TableName != "users" {
		t.Errorf("expected first plan table 'users', got '%s'", result.Plans[0].TableName)
	}

	if result.Plans[1].TableName != "orders" {
		t.Errorf("expected second plan table 'orders', got '%s'", result.Plans[1].TableName)
	}
}

func TestResult_DatabaseTypes(t *testing.T) {
	testCases := []struct {
		dbType     string
		expectType string
	}{
		{"oceanbase_mysql", "oceanbase_mysql"},
		{"mysql", "mysql"},
		{"postgres", "postgres"},
		{"sqlite", "sqlite"},
		{"unknown", "unknown"},
	}

	for _, tc := range testCases {
		result := Result{
			DatabaseType: tc.dbType,
		}

		if result.DatabaseType != tc.expectType {
			t.Errorf("expected DatabaseType '%s', got '%s'", tc.expectType, result.DatabaseType)
		}
	}
}

func TestResult_GlobalRiskLevels(t *testing.T) {
	riskLevels := []string{"low", "medium", "high"}

	for _, risk := range riskLevels {
		result := Result{
			DatabaseType: "mysql",
			GlobalRisk:   risk,
		}

		if result.GlobalRisk != risk {
			t.Errorf("expected GlobalRisk '%s', got '%s'", risk, result.GlobalRisk)
		}
	}
}
