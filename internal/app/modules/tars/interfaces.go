package tars

import (
	"context"

	"assistant/internal/app/modules/tars/memory"
	"assistant/pkg/llm"
)

type IMemoryService interface {
	AddUserMessage(ctx context.Context, chatID, openID, username, content, messageID string) error
	AddAssistantMessage(ctx context.Context, chatID, openID, username, content string) error
	GetContextForLLM(ctx context.Context, chatID string) ([]llm.Message, error)
	GetRecentMessages(ctx context.Context, chatID string) ([]llm.Message, error)
	GetShortTermMessages(chatID string) []memory.ShortTermMessage
	ClearHistory(chatID string) error
	TrySummarizeAndSave(ctx context.Context, chatID string) error
	TryRecallIfNeeded(ctx context.Context, chatID, query string) (string, error)
}
