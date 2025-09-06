package translator

import (
	"testing"
)

func TestNewTranslator(t *testing.T) {
	translator := NewTranslator(nil)

	if translator == nil {
		t.Fatal("expected non-nil Translator even with nil client")
	}
}

func TestResult_Fields(t *testing.T) {
	result := Result{
		SourceLanguage: "English",
		TargetLanguage: "Chinese",
		InputType:      "sentence",
		Translation:    "你好世界",
		Confidence:     0.95,
	}

	if result.SourceLanguage != "English" {
		t.Errorf("expected SourceLanguage 'English', got '%s'", result.SourceLanguage)
	}

	if result.TargetLanguage != "Chinese" {
		t.Errorf("expected TargetLanguage 'Chinese', got '%s'", result.TargetLanguage)
	}

	if result.InputType != "sentence" {
		t.Errorf("expected InputType 'sentence', got '%s'", result.InputType)
	}

	if result.Translation != "你好世界" {
		t.Errorf("unexpected Translation: %s", result.Translation)
	}

	if result.Confidence != 0.95 {
		t.Errorf("expected Confidence 0.95, got %f", result.Confidence)
	}
}

func TestResult_InputTypes(t *testing.T) {
	inputTypes := []string{"word", "sentence", "article"}

	for _, inputType := range inputTypes {
		result := Result{
			InputType: inputType,
		}

		if result.InputType != inputType {
			t.Errorf("expected InputType '%s', got '%s'", inputType, result.InputType)
		}
	}
}

func TestResult_ConfidenceRange(t *testing.T) {
	confidenceValues := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, conf := range confidenceValues {
		result := Result{
			Confidence: conf,
		}

		if result.Confidence != conf {
			t.Errorf("expected Confidence %f, got %f", conf, result.Confidence)
		}
	}
}

func TestResult_CommonLanguages(t *testing.T) {
	testCases := []struct {
		source string
		target string
	}{
		{"English", "Chinese"},
		{"Chinese", "English"},
		{"Japanese", "Chinese"},
		{"Korean", "English"},
		{"French", "English"},
		{"unknown", "English"},
	}

	for _, tc := range testCases {
		result := Result{
			SourceLanguage: tc.source,
			TargetLanguage: tc.target,
		}

		if result.SourceLanguage != tc.source {
			t.Errorf("expected SourceLanguage '%s', got '%s'", tc.source, result.SourceLanguage)
		}

		if result.TargetLanguage != tc.target {
			t.Errorf("expected TargetLanguage '%s', got '%s'", tc.target, result.TargetLanguage)
		}
	}
}

func TestResult_WordTranslation(t *testing.T) {
	result := Result{
		SourceLanguage: "English",
		TargetLanguage: "Chinese",
		InputType:      "word",
		Translation:    "计算机",
		Confidence:     0.98,
	}

	if result.InputType != "word" {
		t.Errorf("expected InputType 'word', got '%s'", result.InputType)
	}

	if len(result.Translation) == 0 {
		t.Error("expected non-empty translation for word")
	}
}

func TestResult_ArticleTranslation(t *testing.T) {
	result := Result{
		SourceLanguage: "English",
		TargetLanguage: "Chinese",
		InputType:      "article",
		Translation:    "# 标题\n\n这是一个段落。",
		Confidence:     0.85,
	}

	if result.InputType != "article" {
		t.Errorf("expected InputType 'article', got '%s'", result.InputType)
	}

	if len(result.Translation) == 0 {
		t.Error("expected non-empty translation for article")
	}
}
