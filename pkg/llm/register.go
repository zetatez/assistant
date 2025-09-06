package llm

import "fmt"

type Factory func(cfg Config) (Client, error)

var providers = map[string]Factory{}

func Register(name string, f Factory) {
	providers[name] = f
}

func NewClient(provider string, cfg Config) (Client, error) {
	f, ok := providers[provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
	return f(cfg)
}
