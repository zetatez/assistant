package memory

import (
	"context"
	"database/sql"
	"strings"
	"sync"

	"assistant/internal/app/repo"
	"assistant/pkg/llm"
)

type Recall struct {
	repo      *repo.Queries
	llmClient llm.Client
	llmModel  string
	logger    Logger
}

func NewRecall(repo *repo.Queries, llmClient llm.Client, llmModel string, logger Logger) *Recall {
	return &Recall{
		repo:      repo,
		llmClient: llmClient,
		llmModel:  llmModel,
		logger:    logger,
	}
}

func (r *Recall) SearchOldMessages(ctx context.Context, sessionID string, keywords []string, limit int) ([]ShortTermMessage, error) {
	if len(keywords) == 0 {
		return nil, nil
	}

	type result struct {
		msgs []ShortTermMessage
	}
	results := make([]result, len(keywords))
	var wg sync.WaitGroup

	for i, kw := range keywords {
		wg.Add(1)
		go func(idx int, pattern string) {
			defer wg.Done()
			messages, err := r.repo.SearchOldChatMessagesByKeyword(ctx, repo.SearchOldChatMessagesByKeywordParams{
				SessionID: sessionID,
				Content:   pattern,
				Limit:     int32(limit),
			})
			if err != nil {
				r.logger.Warnf("recall: search old messages error: %v", err)
				return
			}
			rs := make([]ShortTermMessage, 0, len(messages))
			for _, msg := range messages {
				rs = append(rs, ShortTermMessage{
					ID:      msg.ID,
					Role:    msg.Role,
					Content: msg.Content,
					Time:    msg.CreatedAt,
				})
			}
			results[idx] = result{msgs: rs}
		}(i, "%"+kw+"%")
	}
	wg.Wait()

	seen := make(map[int64]bool)
	var allMsgs []ShortTermMessage
	for _, res := range results {
		for _, msg := range res.msgs {
			if !seen[msg.ID] {
				seen[msg.ID] = true
				allMsgs = append(allMsgs, msg)
			}
		}
	}
	return allMsgs, nil
}

func (r *Recall) RecallByLLM(ctx context.Context, sessionID, query string, oldMessages []ShortTermMessage) (string, error) {
	if r.llmClient == nil || len(oldMessages) == 0 {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("User is asking about something that might be in their old conversation history.\n\n")
	sb.WriteString("Old conversation history:\n")
	for _, msg := range oldMessages {
		sb.WriteString(msg.Role)
		sb.WriteString(": ")
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}
	sb.WriteString("\n---\nUser's current question: ")
	sb.WriteString(query)
	sb.WriteString("\n\nBased on the old conversation history above, if relevant, provide a brief recollection (2-3 sentences) that helps answer the user's current question. If the history is not relevant, respond with only the word: NOT_RELEVANT")

	resp, err := r.llmClient.Chat(ctx, llm.ChatRequest{
		Model:       r.llmModel,
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: sb.String()}},
		Temperature: 0.3,
		MaxTokens:   200,
	})
	if err != nil {
		return "", err
	}

	content := strings.TrimSpace(resp.Content)
	if content == "NOT_RELEVANT" {
		return "", nil
	}

	r.saveRecall(ctx, sessionID, query, content)

	return content, nil
}

func (r *Recall) saveRecall(ctx context.Context, sessionID, query, content string) error {
	_, err := r.repo.CreateChatRecall(ctx, repo.CreateChatRecallParams{
		SessionID:       sessionID,
		Query:           query,
		RecalledContent: content,
		RelevanceScore:  sql.NullFloat64{Float64: 0.8, Valid: true},
	})
	return err
}

func (r *Recall) GetRecentRecalls(ctx context.Context, sessionID, query string, limit int) ([]RecallResult, error) {
	pattern := "%" + query + "%"
	recalls, err := r.repo.SearchChatRecall(ctx, repo.SearchChatRecallParams{
		SessionID: sessionID,
		Query:     pattern,
		Limit:     int32(limit),
	})
	if err != nil {
		return nil, err
	}

	results := make([]RecallResult, len(recalls))
	for i, recall := range recalls {
		results[i] = RecallResult{
			Content:        recall.RecalledContent,
			RelevanceScore: float64(recall.RelevanceScore.Float64),
			SourceTime:     recall.CreatedAt,
		}
	}
	return results, nil
}

func formatMessagesForRecall(messages []ShortTermMessage) string {
	var sb strings.Builder
	for _, msg := range messages {
		sb.WriteString(msg.Time.Format("2006-01-02 15:04"))
		sb.WriteString(" [")
		sb.WriteString(msg.Role)
		sb.WriteString("]: ")
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}
