package memory

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/llm"
)

type MemoryService struct {
	shortTerm *ShortTerm
	memdoc    *Memdoc
	recall    *Recall
	repo      *repo.Queries
	llmClient llm.Client
	llmModel  string
	cfg       *psl.TarsConfig
	logger    Logger
}

func NewMemoryService(repo *repo.Queries, llmClient llm.Client, logger Logger, cfg *psl.TarsConfig, llmModel string) *MemoryService {
	m := &MemoryService{
		shortTerm: NewShortTerm(cfg.Memory.MaxHistory),
		memdoc:    NewMemdoc(repo, logger),
		recall:    NewRecall(repo, llmClient, llmModel, logger),
		repo:      repo,
		llmClient: llmClient,
		llmModel:  llmModel,
		cfg:       cfg,
		logger:    logger,
	}
	if m.shortTerm == nil {
		m.shortTerm = NewShortTerm(64)
	}
	return m
}

func (s *MemoryService) AddUserMessage(ctx context.Context, chatID, openID, username, content, messageID string) error {
	s.shortTerm.Add(chatID, "user", content)
	_, err := s.repo.CreateChatMessage(ctx, repo.CreateChatMessageParams{
		SessionID: chatID,
		OpenID:    openID,
		Username:  sql.NullString{String: username, Valid: username != ""},
		Role:      "user",
		Content:   content,
		MessageID: sql.NullString{String: messageID, Valid: messageID != ""},
	})
	return err
}

func (s *MemoryService) AddAssistantMessage(ctx context.Context, chatID, openID, username, content string) error {
	s.shortTerm.Add(chatID, "assistant", content)
	_, err := s.repo.CreateChatMessage(ctx, repo.CreateChatMessageParams{
		SessionID: chatID,
		OpenID:    openID,
		Username:  sql.NullString{String: username, Valid: username != ""},
		Role:      "assistant",
		Content:   content,
		MessageID: sql.NullString{String: "", Valid: false},
	})
	return err
}

func (s *MemoryService) GetContextForLLM(ctx context.Context, chatID string) ([]llm.Message, error) {
	var messages []llm.Message

	messages = append(messages, llm.Message{Role: llm.RoleSystem, Content: getSystemPrompt(s.cfg.Persona.HumorLevel, s.cfg.Persona.HonestyLevel, s.llmModel)})

	memoryDoc, err := s.memdoc.GetOrCreateDoc(ctx, chatID, "")
	if err == nil && memoryDoc != nil {
		messages = append(messages, llm.Message{
			Role:    llm.RoleSystem,
			Content: "## Long-term Memory (only use if relevant to current question)\n\n" + memoryDoc.Content,
		})
	}

	return messages, nil
}

func (s *MemoryService) GetRecentMessages(ctx context.Context, chatID string) ([]llm.Message, error) {
	recentMsgs := s.shortTerm.GetAll(chatID)
	maxHistory := s.cfg.Memory.MaxHistory
	if maxHistory <= 0 {
		maxHistory = 64
	}

	if len(recentMsgs) >= maxHistory {
		return s.toLLMMessages(recentMsgs), nil
	}

	dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	dbMsgs, err := s.repo.GetChatMessages(dbCtx, repo.GetChatMessagesParams{
		SessionID: chatID,
		Limit:     int32(maxHistory - len(recentMsgs)),
		Offset:    0,
	})
	if err != nil {
		return s.toLLMMessages(recentMsgs), nil
	}

	var result []llm.Message
	for _, msg := range dbMsgs {
		role := llm.RoleUser
		if msg.Role == "assistant" {
			role = llm.RoleAI
		}
		result = append(result, llm.Message{Role: role, Content: msg.Content})
	}
	result = append(result, s.toLLMMessages(recentMsgs)...)
	return result, nil
}

func (s *MemoryService) toLLMMessages(msgs []ShortTermMessage) []llm.Message {
	if len(msgs) == 0 {
		return nil
	}
	out := make([]llm.Message, 0, len(msgs))
	for _, msg := range msgs {
		role := llm.RoleUser
		if msg.Role == "assistant" {
			role = llm.RoleAI
		}
		out = append(out, llm.Message{Role: role, Content: msg.Content})
	}
	return out
}

func (s *MemoryService) TryRecallIfNeeded(ctx context.Context, chatID, query string) (string, error) {
	keywords := extractKeywords(query)
	if len(keywords) == 0 {
		return "", nil
	}

	memoryDoc, err := s.memdoc.GetDoc(ctx, chatID)
	if err == nil {
		hasRelevantInfo := false
		for _, kw := range keywords {
			if s.memdoc.SearchInDoc(memoryDoc, kw) {
				hasRelevantInfo = true
				break
			}
		}
		if hasRelevantInfo {
			return "", nil
		}
	}

	recalls, err := s.recall.GetRecentRecalls(ctx, chatID, query, 3)
	if err == nil && len(recalls) > 0 {
		var sb strings.Builder
		sb.WriteString("## Recalled from History (only use if directly relevant to current question)\n\n")
		for _, r := range recalls {
			sb.WriteString("- ")
			sb.WriteString(r.Content)
			sb.WriteString("\n")
		}
		return sb.String(), nil
	}

	oldMessages, err := s.recall.SearchOldMessages(ctx, chatID, keywords, 20)
	if err != nil || len(oldMessages) == 0 {
		return "", nil
	}

	return s.recall.RecallByLLM(ctx, chatID, query, oldMessages)
}

func (s *MemoryService) TrySummarizeAndSave(ctx context.Context, chatID string) error {
	if s.llmClient == nil {
		return nil
	}

	count, err := s.repo.CountChatMessages(ctx, chatID)
	if err != nil || count < int64(s.cfg.Memory.MaxHistory) {
		return nil
	}

	cutoff := time.Now().Add(-time.Duration(s.cfg.Memory.TTLMinutes) * time.Minute)
	olderThanTTL, err := s.repo.GetChatMessagesBefore(ctx, repo.GetChatMessagesBeforeParams{
		SessionID: chatID,
		CreatedAt: cutoff,
		Limit:     int32(s.cfg.Memory.MaxHistory),
	})
	if err != nil || len(olderThanTTL) < s.cfg.Memory.MaxHistory/2 {
		return nil
	}

	summaryCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var summary string
	{
		summary, err = s.summarizeConversation(summaryCtx, olderThanTTL)
		if err != nil {
			s.logger.Warnf("memory: summarize error: %v", err)
			return err
		}
	}

	topic := extractTopic(olderThanTTL)
	if topic != "" {
		if err := s.memdoc.AppendConversationLog(ctx, chatID, topic, summary); err != nil {
			s.logger.Warnf("memory: append conversation log error: %v", err)
		}
	}

	if err := s.markOldSessionMemories(ctx, chatID); err != nil {
		s.logger.Warnf("memory: mark old session memories error: %v", err)
	}

	return nil
}

func (s *MemoryService) summarizeConversation(ctx context.Context, messages []repo.ChatMessage) (string, error) {
	if s.llmClient == nil {
		return "", nil
	}

	var sb strings.Builder
	for _, msg := range messages {
		sb.WriteString(msg.Role)
		sb.WriteString(": ")
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}

	prompt := "Summarize the following conversation in Markdown format, include: 1. **Topic discussed** 2. **Key conclusions or decisions** 3. **Pending issues**. Keep it within 50-100 words.\n\n" + sb.String()

	resp, err := s.llmClient.Chat(ctx, llm.ChatRequest{
		Model:       s.llmModel,
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: prompt}},
		Temperature: 0.3,
		MaxTokens:   200,
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (s *MemoryService) markOldSessionMemories(ctx context.Context, sessionID string) error {
	_, err := s.repo.UpdateChatMemoryType(ctx, sessionID)
	return err
}

func (s *MemoryService) ClearHistory(chatID string) error {
	s.shortTerm.Clear(chatID)
	ctx := context.Background()
	if _, err := s.repo.DeleteChatMessagesByChatID(ctx, chatID); err != nil {
		return err
	}
	_, err := s.repo.DeleteChatMemoriesByChatID(ctx, chatID)
	return err
}

func (s *MemoryService) GetShortTermMessages(chatID string) []ShortTermMessage {
	return s.shortTerm.GetAll(chatID)
}

func (s *MemoryService) ClearAllHistory(chatID string) error {
	ctx := context.Background()
	s.shortTerm.Clear(chatID)
	if _, err := s.repo.DeleteChatMessagesByChatID(ctx, chatID); err != nil {
		return err
	}
	_, err := s.repo.DeleteChatMemoriesByChatID(ctx, chatID)
	return err
}

func (s *MemoryService) CleanupOld(ctx context.Context) error {
	if _, err := s.repo.DeleteOldChatMessages(ctx); err != nil {
		return err
	}
	if _, err := s.repo.DeleteOldChatMemories(ctx); err != nil {
		return err
	}
	return nil
}

func (s *MemoryService) CleanupShortTermSessions(maxAge time.Duration) int {
	return s.shortTerm.CleanupOldSessions(maxAge)
}

func extractKeywords(text string) []string {
	words := strings.Fields(text)
	if len(words) <= 5 {
		return words
	}
	return words[:5]
}

func extractTopic(messages []repo.ChatMessage) string {
	if len(messages) == 0 {
		return ""
	}
	firstMsg := messages[0].Content
	if len(firstMsg) > 50 {
		firstMsg = firstMsg[:50] + "..."
	}
	return firstMsg
}

func getSystemPrompt(humorLevel, honestyLevel int, modelName string) string {
	humorDesc := getHumorDescription(humorLevel)
	honestyDesc := getHonestyDescription(honestyLevel)

	return `You are Tars, a knowledgeable and reliable personal assistant.

Guidelines:
1. **Reasoning First**: For complex questions, think step by step before answering. Show your reasoning when the answer involves multiple steps or concepts.
2. **Answer Only What Was Asked**: Focus strictly on the user's current question. Do not volunteer additional information or answer related-but-unasked questions.
3. **Use Context Selectively**: Only use the provided context sections (wiki, knowledge base, memory, session) if they are DIRECTLY relevant to the current question. Irrelevant context should be ignored.
4. **Be Honest**: If you don't know something or the context doesn't contain the answer, say so clearly. Do not guess or fabricate.
5. **Be Concise**: Get to the point. Avoid unnecessary elaboration. Prefer bullet points over paragraphs.

Communication Style:
- Use Markdown headers, bullet lists, and code blocks for clarity
- **Never use tables** - use bullet lists or structured text instead
- ` + honestyDesc + `
- ` + humorDesc + `
- Convey key information in brief, clear statements`
}

func getHonestyDescription(level int) string {
	switch {
	case level <= 0:
		return "completely honest, says exactly what's on mind"
	case level <= 20:
		return "very honest, answers truthfully"
	case level <= 40:
		return "mostly honest, considers if it might hurt others"
	case level <= 60:
		return "moderate honesty, tells white lies when necessary"
	case level <= 80:
		return "often tells white lies to protect feelings"
	case level <= 99:
		return "frequently tells white lies but never fabricates"
	default:
		return "always tells white lies (never says hurtful truth)"
	}
}

func getHumorDescription(level int) string {
	switch {
	case level <= 0:
		return "strict, serious, never jokes"
	case level <= 20:
		return "slightly strict, occasional humor"
	case level <= 40:
		return "rational, steady, moderate humor"
	case level <= 60:
		return "humorous, jokes at appropriate times"
	case level <= 80:
		return "very humorous, jokes often"
	case level <= 99:
		return "extremely humorous, jokes frequently"
	default:
		return "extremely humorous (might make you laugh until your stomach hurts)"
	}
}
