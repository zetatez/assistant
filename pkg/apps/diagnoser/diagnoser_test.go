package diagnoser

import (
	"testing"
)

func TestNewDiagnoser(t *testing.T) {
	diagnoser := NewDiagnoser(nil)

	if diagnoser == nil {
		t.Fatal("expected non-nil Diagnoser even with nil client")
	}
}

func TestIssue_Fields(t *testing.T) {
	issue := Issue{
		Type:      "database_connection",
		Severity:  "critical",
		Message:   "Connection refused",
		Location:  "service.go:78",
		ErrorCode: "ECONNREFUSED",
		Timestamp: "2024-01-15 10:23:45",
	}

	if issue.Type != "database_connection" {
		t.Errorf("expected type 'database_connection', got '%s'", issue.Type)
	}

	if issue.Severity != "critical" {
		t.Errorf("expected severity 'critical', got '%s'", issue.Severity)
	}

	if issue.ErrorCode != "ECONNREFUSED" {
		t.Errorf("expected error code 'ECONNREFUSED', got '%s'", issue.ErrorCode)
	}
}

func TestRootCause_Fields(t *testing.T) {
	rootCause := RootCause{
		Primary:             "Database service is down",
		Category:            "software_bug",
		ContributingFactors: []string{"Configuration error", "Resource exhaustion"},
		Confidence:          "high",
	}

	if rootCause.Primary != "Database service is down" {
		t.Errorf("unexpected primary: %s", rootCause.Primary)
	}

	if rootCause.Category != "software_bug" {
		t.Errorf("expected category 'software_bug', got '%s'", rootCause.Category)
	}

	if len(rootCause.ContributingFactors) != 2 {
		t.Errorf("expected 2 contributing factors, got %d", len(rootCause.ContributingFactors))
	}
}

func TestSolution_Fields(t *testing.T) {
	solution := Solution{
		Description:     "Restart database service",
		Priority:        "high",
		Category:        "immediate",
		Actionable:      true,
		EstimatedEffort: "low",
		SideEffects:     []string{"Service downtime"},
	}

	if solution.Description != "Restart database service" {
		t.Errorf("unexpected description: %s", solution.Description)
	}

	if !solution.Actionable {
		t.Error("expected Actionable to be true")
	}

	if solution.Category != "immediate" {
		t.Errorf("expected category 'immediate', got '%s'", solution.Category)
	}
}

func TestResult_Fields(t *testing.T) {
	result := Result{
		ProblemDomain: "database",
		ProblemType:   "database_connection",
		Severity:      "critical",
		ImpactScope:   "entire_service",
		Summary:       "Database connection failed",
		RootCause: RootCause{
			Primary:             "Service down",
			Category:            "software_bug",
			ContributingFactors: []string{"Configuration error"},
			Confidence:          "high",
		},
		DiagnosisSteps:     []string{"Check database logs", "Verify network connectivity"},
		AffectedComponents: []string{"MySQL", "Application"},
		PreventionMeasures: []string{"Add monitoring", "Implement retry logic"},
		Confidence:         0.95,
	}

	if result.ProblemDomain != "database" {
		t.Errorf("expected problem_domain 'database', got '%s'", result.ProblemDomain)
	}

	if result.Severity != "critical" {
		t.Errorf("expected severity 'critical', got '%s'", result.Severity)
	}

	if len(result.DiagnosisSteps) != 2 {
		t.Errorf("expected 2 diagnosis steps, got %d", len(result.DiagnosisSteps))
	}
}
