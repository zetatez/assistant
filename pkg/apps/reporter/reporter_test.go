package reporter

import (
	"testing"
)

func TestNewReporter(t *testing.T) {
	reporter := NewReporter(nil)

	if reporter == nil {
		t.Fatal("expected non-nil Reporter even with nil client")
	}
}

func TestReportType_Constants(t *testing.T) {
	if WeeklyReport != "weekly" {
		t.Errorf("expected WeeklyReport to be 'weekly', got '%s'", WeeklyReport)
	}

	if MonthlyReport != "monthly" {
		t.Errorf("expected MonthlyReport to be 'monthly', got '%s'", MonthlyReport)
	}

	if QuarterReport != "quarterly" {
		t.Errorf("expected QuarterReport to be 'quarterly', got '%s'", QuarterReport)
	}

	if YearlyReport != "yearly" {
		t.Errorf("expected YearlyReport to be 'yearly', got '%s'", YearlyReport)
	}
}

func TestInput_Fields(t *testing.T) {
	input := Input{
		ReportType:  WeeklyReport,
		Author:      "John Doe",
		Role:        "Software Engineer",
		Period:      "2026-01-06 ~ 2026-01-12",
		Language:    "zh-CN",
		WorkContent: "Completed feature development",
	}

	if input.ReportType != WeeklyReport {
		t.Errorf("expected ReportType 'weekly', got '%s'", input.ReportType)
	}

	if input.Author != "John Doe" {
		t.Errorf("expected Author 'John Doe', got '%s'", input.Author)
	}

	if input.Role != "Software Engineer" {
		t.Errorf("expected Role 'Software Engineer', got '%s'", input.Role)
	}

	if input.Period != "2026-01-06 ~ 2026-01-12" {
		t.Errorf("expected Period '2026-01-06 ~ 2026-01-12', got '%s'", input.Period)
	}

	if input.Language != "zh-CN" {
		t.Errorf("expected Language 'zh-CN', got '%s'", input.Language)
	}

	if input.WorkContent != "Completed feature development" {
		t.Errorf("expected WorkContent 'Completed feature development', got '%s'", input.WorkContent)
	}
}

func TestResult_Fields(t *testing.T) {
	result := Result{
		FileName:   "工作周报-2026-01-06.2026-01-12.md",
		ReportType: "weekly",
		Language:   "zh-CN",
		Markdown:   "# 工作周报\n\n本周完成...",
		Confidence: 0.92,
	}

	if result.FileName != "工作周报-2026-01-06.2026-01-12.md" {
		t.Errorf("unexpected file name: %s", result.FileName)
	}

	if result.ReportType != "weekly" {
		t.Errorf("expected ReportType 'weekly', got '%s'", result.ReportType)
	}

	if result.Language != "zh-CN" {
		t.Errorf("expected Language 'zh-CN', got '%s'", result.Language)
	}

	if result.Confidence != 0.92 {
		t.Errorf("expected Confidence 0.92, got %f", result.Confidence)
	}

	if result.Markdown == "" {
		t.Error("expected non-empty Markdown")
	}
}
