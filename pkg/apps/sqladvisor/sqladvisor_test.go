package sqladvisor

import (
	"testing"
)

func TestNewSQLAdvisor(t *testing.T) {
	advisor := NewSQLAdvisor(nil)

	if advisor == nil {
		t.Fatal("expected non-nil SQLAdvisor even with nil client")
	}
}

func TestResult_Fields(t *testing.T) {
	result := Result{
		DatabaseType:  "mysql",
		OriginalSQL:   "SELECT * FROM users WHERE name = 'John'",
		OptimizedSQL:  "SELECT id, name, email FROM users WHERE name = 'John'",
		Optimizations: []string{"Avoid SELECT *", "Explicit column list"},
		RiskLevel:     "low",
		Confidence:    0.95,
	}

	if result.DatabaseType != "mysql" {
		t.Errorf("expected DatabaseType 'mysql', got '%s'", result.DatabaseType)
	}

	if result.OriginalSQL != "SELECT * FROM users WHERE name = 'John'" {
		t.Errorf("unexpected OriginalSQL: %s", result.OriginalSQL)
	}

	if result.OptimizedSQL != "SELECT id, name, email FROM users WHERE name = 'John'" {
		t.Errorf("unexpected OptimizedSQL: %s", result.OptimizedSQL)
	}

	if len(result.Optimizations) != 2 {
		t.Errorf("expected 2 optimizations, got %d", len(result.Optimizations))
	}

	if result.RiskLevel != "low" {
		t.Errorf("expected RiskLevel 'low', got '%s'", result.RiskLevel)
	}

	if result.Confidence != 0.95 {
		t.Errorf("expected Confidence 0.95, got %f", result.Confidence)
	}
}

func TestResult_EmptyOptimizations(t *testing.T) {
	result := Result{
		DatabaseType:  "mysql",
		OriginalSQL:   "SELECT id FROM users WHERE id = 1",
		OptimizedSQL:  "SELECT id FROM users WHERE id = 1",
		Optimizations: []string{},
		RiskLevel:     "low",
		Confidence:    1.0,
	}

	if len(result.Optimizations) != 0 {
		t.Errorf("expected 0 optimizations, got %d", len(result.Optimizations))
	}

	if result.Confidence != 1.0 {
		t.Errorf("expected Confidence 1.0, got %f", result.Confidence)
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

func TestResult_RiskLevels(t *testing.T) {
	riskLevels := []string{"low", "medium", "high"}

	for _, risk := range riskLevels {
		result := Result{
			DatabaseType: "mysql",
			RiskLevel:    risk,
		}

		if result.RiskLevel != risk {
			t.Errorf("expected RiskLevel '%s', got '%s'", risk, result.RiskLevel)
		}
	}
}
