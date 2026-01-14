package llm

type Config struct {
	APIKey  string
	BaseURL string
	Model   string
	Extra   map[string]string
}
