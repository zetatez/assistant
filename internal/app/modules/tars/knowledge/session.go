package knowledge

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"assistant/internal/app/repo"
	"assistant/pkg/llm"
)

type SessionManager struct {
	repo      *repo.Queries
	llmClient llm.Client
	llmModel  string
	logger    Logger
}

type SessionState struct {
	SessionID    string
	Summary      string
	PendingTasks []PendingTask
	Context      string
	UpdatedAt    time.Time
}

// SessionState vs LongTermMemory distinction:
// - SessionState (chat_session): Current session, updated on every message
//   Stores: summary, context, pending_tasks - real-time current conversation state
// - LongTermMemory (chat_memory): Historical summaries, created periodically
//   Stores: historical conversation summaries for long-term retrieval
// Both store "summary" but serve different purposes:
//   SessionState.summary = current, real-time understanding
//   chat_memory.summary = historical snapshots for memory retrieval

type PendingTask struct {
	Content   string `json:"content"`
	DueTime   string `json:"due_time,omitempty"`
	Priority  string `json:"priority,omitempty"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func NewSessionManager(repo *repo.Queries, llmClient llm.Client, llmModel string, logger Logger) *SessionManager {
	return &SessionManager{
		repo:      repo,
		llmClient: llmClient,
		llmModel:  llmModel,
		logger:    logger,
	}
}

func (s *SessionManager) GetSession(ctx context.Context, sessionID string) (*SessionState, error) {
	session, err := s.repo.GetOrCreateChatSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	state := &SessionState{
		SessionID: session.SessionID,
		Summary:   session.Summary.String,
		Context:   session.Context.String,
		UpdatedAt: time.Time{},
	}
	if session.UpdatedAt.Valid {
		state.UpdatedAt = session.UpdatedAt.Time
	}

	if session.PendingTasks.String != "" {
		if err := json.Unmarshal([]byte(session.PendingTasks.String), &state.PendingTasks); err != nil {
			state.PendingTasks = nil
		}
	}

	return state, nil
}

func (s *SessionManager) UpdateSession(ctx context.Context, sessionID, summary, context string, tasks []PendingTask) error {
	tasksJSON, _ := json.Marshal(tasks)
	_, err := s.repo.UpsertChatSession(ctx, repo.UpsertChatSessionParams{
		SessionID:    sessionID,
		Summary:      sql.NullString{String: summary, Valid: summary != ""},
		PendingTasks: sql.NullString{String: string(tasksJSON), Valid: len(tasksJSON) > 2},
		Context:      sql.NullString{String: context, Valid: context != ""},
	})
	return err
}

func (s *SessionManager) RefreshSessionSummary(ctx context.Context, chatID string, recentMessages []string) error {
	if s.llmClient == nil || len(recentMessages) == 0 {
		return nil
	}

	var sb strings.Builder
	for _, msg := range recentMessages {
		sb.WriteString(msg)
		sb.WriteString("\n")
	}

	prompt := `Based on the recent conversation below, update the session summary.
Focus on: important context, key decisions, pending tasks, discussion progress.
Keep it concise, 50-100 words.

Recent conversation:
` + sb.String() + `

Respond with a JSON object:
{"summary": "...", "context": "...", "tasks": [{"content": "...", "status": "pending"}]}`

	resp, err := s.llmClient.Chat(ctx, llm.ChatRequest{
		Model:       s.llmModel,
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: prompt}},
		Temperature: 0.3,
		MaxTokens:   500,
	})
	if err != nil {
		return err
	}

	var result struct {
		Summary string        `json:"summary"`
		Context string        `json:"context"`
		Tasks   []PendingTask `json:"tasks"`
	}

	content := strings.TrimSpace(resp.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		s.logger.Warnf("knowledge: failed to parse session summary response: %v, content: %s", err, content)
		if err := s.parseSessionSummaryFallback(content, &result); err != nil {
			return err
		}
	}

	return s.UpdateSession(ctx, chatID, result.Summary, result.Context, result.Tasks)
}

func (s *SessionManager) parseSessionSummaryFallback(content string, result *struct {
	Summary string        `json:"summary"`
	Context string        `json:"context"`
	Tasks   []PendingTask `json:"tasks"`
}) error {
	result.Summary = content
	result.Context = ""
	result.Tasks = nil
	return nil
}
