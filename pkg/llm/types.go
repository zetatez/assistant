package llm

type Role string

const (
	RoleSystem Role = "system"
	RoleUser   Role = "user"
	RoleAI     Role = "assistant"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string
	Messages    []Message
	Temperature float32
	MaxTokens   int
}

type ChatResponse struct {
	Content string
	Usage   TokenUsage
}

type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type StreamCallback func(delta string)

type Capability uint64

const (
	CapabilityChat Capability = 1 << iota
	CapabilityStream
	CapabilityFunctionCall
	CapabilityVision
)

type Capabilities struct {
	Supported Capability
}

func (c Capabilities) Has(cap Capability) bool {
	return c.Supported&cap != 0
}
