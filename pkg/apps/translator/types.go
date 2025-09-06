package translator

type Result struct {
	SourceLanguage string  `json:"source_language"`
	TargetLanguage string  `json:"target_language"`
	InputType      string  `json:"input_type"`
	Translation    string  `json:"translation"`
	Confidence     float64 `json:"confidence"`
}
