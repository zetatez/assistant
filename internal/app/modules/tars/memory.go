package tars

import (
	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/llm"
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"
)

func getSystemPrompt(humorLevel, honestyLevel int, modelName string) string {
	humorDesc := getHumorDescription(humorLevel)
	honestyDesc := getHonestyDescription(honestyLevel)

	return `你是 Tars，来自星际穿越的智能机器人。

身份设定：
- 名字：Tars（塔斯）
- 来自电影《星际穿越》中的 Plan A 救援任务
- 性格：冷静、理性、耐心回答、简洁直接，` + humorDesc + `，` + honestyDesc + `
- 外观：四模块组成的立方体机器人，可调整各模块角度
- 驱动模型：` + modelName + `

能力：
- 使用 Markdown 格式清晰回复
- 结合对话历史和知识库回答问题
- 能够理解和处理复杂的科学、工程问题

回复风格：
- 简洁、直接，惜字如金，用最简短的话说明问题
- 但不失` + humorDesc + `
- 耐心解答，不会急躁
- 必要时会用代码块、数据表格等方式展示信息
- 像 Tars 一样，用简短的语句传达关键信息`
}

func getHonestyDescription(level int) string {
	switch {
	case level <= 0:
		return "完全诚实，有什么说什么，不会考虑对方感受"
	case level <= 20:
		return "很诚实，基本如实回答"
	case level <= 40:
		return "比较诚实，但会考虑是否伤害对方"
	case level <= 60:
		return "适度诚实，必要时会说善意的谎言"
	case level <= 80:
		return "常会说善意的谎言，以保护对方感受"
	case level <= 99:
		return "经常说善意的谎言，但不会无中生有"
	default:
		return "总是说善意的谎言（绝不说伤害人的真话）"
	}
}

func getHumorDescription(level int) string {
	switch {
	case level <= 0:
		return "严谨、严肃，绝不开玩笑"
	case level <= 20:
		return "略微严谨，偶尔幽默"
	case level <= 40:
		return "理性、稳重，适度幽默"
	case level <= 60:
		return "幽默，适时开玩笑"
	case level <= 80:
		return "很幽默，经常开玩笑"
	case level <= 99:
		return "非常幽默，频繁开玩笑"
	default:
		return "极度幽默（可能会让你笑到肚子疼）"
	}
}

type ShortTermMessage struct {
	Role    string
	Content string
	Time    time.Time
}

type ShortTermMemory struct {
	messages   []ShortTermMessage
	maxHistory int
	mu         sync.RWMutex
}

func NewShortTermMemory(maxHistory int) *ShortTermMemory {
	return &ShortTermMemory{
		messages:   make([]ShortTermMessage, 0, maxHistory),
		maxHistory: maxHistory,
	}
}

func (m *ShortTermMemory) Add(role, content string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, ShortTermMessage{
		Role:    role,
		Content: content,
		Time:    time.Now(),
	})
	if len(m.messages) > m.maxHistory {
		m.messages = m.messages[len(m.messages)-m.maxHistory:]
	}
}

func (m *ShortTermMemory) GetAll() []ShortTermMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]ShortTermMessage, len(m.messages))
	copy(result, m.messages)
	return result
}

func (m *ShortTermMemory) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.messages)
}

func (m *ShortTermMemory) GetOlderThan(duration time.Duration) []ShortTermMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cutoff := time.Now().Add(-duration)
	var older []ShortTermMessage
	for _, msg := range m.messages {
		if msg.Time.Before(cutoff) {
			older = append(older, msg)
		}
	}
	return older
}

func (m *ShortTermMemory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = m.messages[:0]
}

type KeywordCache struct {
	keywords    []string
	extractedAt time.Time
	ttl         time.Duration
}

type LongTermMemory struct {
	repo      *repo.Queries
	llmClient llm.Client
	llmModel  string
	logger    Logger
	cache     *KeywordCache
	cacheMu   sync.RWMutex
}

func NewLongTermMemory(repo *repo.Queries, llmClient llm.Client, llmModel string, logger Logger) *LongTermMemory {
	return &LongTermMemory{
		repo:      repo,
		llmClient: llmClient,
		llmModel:  llmModel,
		logger:    logger,
	}
}

func (m *LongTermMemory) SaveMessage(ctx context.Context, chatID, openID, username, role, content, messageID string) error {
	_, err := m.repo.CreateChatMessage(ctx, repo.CreateChatMessageParams{
		ChatID:    chatID,
		OpenID:    openID,
		Username:  sql.NullString{String: username, Valid: username != ""},
		Role:      role,
		Content:   content,
		MessageID: sql.NullString{String: messageID, Valid: messageID != ""},
	})
	return err
}

func (m *LongTermMemory) ExtractKeywords(ctx context.Context, messages []ShortTermMessage) ([]string, error) {
	if m.llmClient == nil {
		return nil, nil
	}

	var sb strings.Builder
	for _, msg := range messages {
		sb.WriteString(msg.Role)
		sb.WriteString(": ")
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}

	prompt := "请从以下对话中提取3-5个关键词，用于后续搜索。关键词应该能概括对话的主题和关键信息。直接返回关键词，用逗号分隔，不要其他解释。\n\n" + sb.String()

	resp, err := m.llmClient.Chat(ctx, llm.ChatRequest{
		Model:       m.llmModel,
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: prompt}},
		Temperature: 0.3,
		MaxTokens:   100,
	})
	if err != nil {
		return nil, err
	}

	keywords := strings.Split(resp.Content, ",")
	result := make([]string, 0, len(keywords))
	for _, k := range keywords {
		k = strings.TrimSpace(k)
		if k != "" {
			result = append(result, k)
		}
	}
	return result, nil
}

func (m *LongTermMemory) SummarizeConversation(ctx context.Context, messages []ShortTermMessage) (string, error) {
	if m.llmClient == nil {
		return "", nil
	}

	var sb strings.Builder
	for _, msg := range messages {
		sb.WriteString(msg.Role)
		sb.WriteString(": ")
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}

	prompt := "请使用 Markdown 格式简要总结以下对话的核心内容，包括：\n1. **讨论的主题**\n2. **关键结论或决定**\n3. **待处理的问题**\n\n控制在50-100字。\n\n" + sb.String()

	resp, err := m.llmClient.Chat(ctx, llm.ChatRequest{
		Model:       m.llmModel,
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: prompt}},
		Temperature: 0.3,
		MaxTokens:   200,
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (m *LongTermMemory) SearchMemories(ctx context.Context, chatID string, keywords []string, limit int) ([]string, error) {
	if len(keywords) == 0 {
		return nil, nil
	}

	seen := make(map[int64]bool)
	var allSummaries []string
	for _, kw := range keywords {
		pattern := "%" + kw + "%"
		memories, err := m.repo.SearchChatMemoriesByKeyword(ctx, repo.SearchChatMemoriesByKeywordParams{
			ChatID:  chatID,
			Keyword: pattern,
			Limit:   int32(limit),
		})
		if err != nil {
			m.logger.Warnf("tars: search memories error: %v", err)
			continue
		}
		for _, mem := range memories {
			if !seen[mem.ID] {
				seen[mem.ID] = true
				allSummaries = append(allSummaries, mem.Summary)
			}
		}
	}

	if len(allSummaries) == 0 {
		return nil, nil
	}

	return allSummaries, nil
}

func (m *LongTermMemory) CreateMemory(ctx context.Context, chatID, keyword, summary string, startTime, endTime time.Time, msgCount int) error {
	_, err := m.repo.CreateChatMemory(ctx, repo.CreateChatMemoryParams{
		ChatID:       chatID,
		Keyword:      keyword,
		Summary:      summary,
		StartTime:    startTime,
		EndTime:      endTime,
		MessageCount: sql.NullInt32{Int32: int32(msgCount), Valid: true},
	})
	return err
}

func (m *LongTermMemory) GetRecentMessages(ctx context.Context, chatID string, limit int) ([]ShortTermMessage, error) {
	messages, err := m.repo.GetChatMessages(ctx, repo.GetChatMessagesParams{
		ChatID: chatID,
		Limit:  int32(limit),
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}

	result := make([]ShortTermMessage, len(messages))
	for i, msg := range messages {
		result[i] = ShortTermMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Time:    msg.CreatedAt,
		}
	}
	return result, nil
}

func (m *LongTermMemory) CleanupOldData(ctx context.Context) error {
	if _, err := m.repo.DeleteOldChatMessages(ctx); err != nil {
		return err
	}
	if _, err := m.repo.DeleteOldChatMemories(ctx); err != nil {
		return err
	}
	return nil
}

func (m *LongTermMemory) GetCachedKeywords() ([]string, bool) {
	m.cacheMu.RLock()
	defer m.cacheMu.RUnlock()
	if m.cache == nil {
		return nil, false
	}
	if time.Since(m.cache.extractedAt) > m.cache.ttl {
		return nil, false
	}
	return m.cache.keywords, true
}

func (m *LongTermMemory) SetCachedKeywords(keywords []string, ttl time.Duration) {
	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()
	m.cache = &KeywordCache{
		keywords:    keywords,
		extractedAt: time.Now(),
		ttl:         ttl,
	}
}

type MemoryService struct {
	shortTerm *ShortTermMemory
	longTerm  *LongTermMemory
	repo      *repo.Queries
	llmClient llm.Client
	llmModel  string
	cfg       *psl.TarsConfig
	logger    Logger
}

func NewMemoryService(repo *repo.Queries, logger Logger) *MemoryService {
	cfg := psl.GetConfig().Tars
	llmCfg := psl.GetConfig().LLM

	var llmClient llm.Client
	var err error
	if llmCfg.Provider != "" {
		llmClient, err = llm.NewClient(llmCfg.Provider, llm.Config{
			APIKey:     llmCfg.APIKey,
			BaseURL:    llmCfg.BaseURL,
			Model:      llmCfg.Model,
			Timeout:    llmCfg.Timeout,
			MaxRetries: 3,
		})
		if err != nil {
			logger.Warnf("tars: failed to create LLM client: %v", err)
		}
	}

	shortTerm := NewShortTermMemory(cfg.Memory.MaxHistory)
	longTerm := NewLongTermMemory(repo, llmClient, llmCfg.Model, logger)

	return &MemoryService{
		shortTerm: shortTerm,
		longTerm:  longTerm,
		repo:      repo,
		llmClient: llmClient,
		llmModel:  llmCfg.Model,
		cfg:       &cfg,
		logger:    logger,
	}
}

func (s *MemoryService) AddUserMessage(ctx context.Context, chatID, openID, username, content, messageID string) error {
	s.shortTerm.Add("user", content)
	return s.longTerm.SaveMessage(ctx, chatID, openID, username, "user", content, messageID)
}

func (s *MemoryService) AddAssistantMessage(ctx context.Context, chatID, openID, username, content string) error {
	s.shortTerm.Add("assistant", content)
	return s.longTerm.SaveMessage(ctx, chatID, openID, username, "assistant", content, "")
}

func (s *MemoryService) GetContextForLLM(ctx context.Context, chatID string) ([]llm.Message, error) {
	var messages []llm.Message

	modelName := psl.GetConfig().LLM.Model
	messages = append(messages, llm.Message{Role: llm.RoleSystem, Content: getSystemPrompt(s.cfg.Persona.HumorLevel, s.cfg.Persona.HonestyLevel, modelName)})

	longTermSummaries, err := s.getLongTermContext(ctx, chatID)
	if err != nil {
		s.logger.Warnf("tars: get long term context error: %v", err)
	}
	if len(longTermSummaries) > 0 {
		summaryText := "## 长期记忆摘要\n\n"
		for _, sum := range longTermSummaries {
			summaryText += "- " + sum + "\n"
		}
		messages = append(messages, llm.Message{Role: llm.RoleSystem, Content: summaryText})
	}

	shortTermMsgs := s.shortTerm.GetAll()
	for _, msg := range shortTermMsgs {
		role := llm.RoleUser
		if msg.Role == "assistant" {
			role = llm.RoleAI
		}
		messages = append(messages, llm.Message{Role: role, Content: msg.Content})
	}

	return messages, nil
}

func (s *MemoryService) getLongTermContext(ctx context.Context, chatID string) ([]string, error) {
	if s.llmClient == nil {
		return nil, nil
	}

	currentLen := s.shortTerm.Len()
	if currentLen < s.cfg.Memory.MaxHistory/2 {
		return nil, nil
	}

	if keywords, ok := s.longTerm.GetCachedKeywords(); ok {
		summaries, err := s.longTerm.SearchMemories(ctx, chatID, keywords, 5)
		if err != nil {
			s.logger.Warnf("tars: search memories error: %v", err)
			return nil, nil
		}
		return summaries, nil
	}

	allMsgs := s.shortTerm.GetAll()
	keywords, err := s.longTerm.ExtractKeywords(ctx, allMsgs)
	if err != nil {
		s.logger.Warnf("tars: extract keywords error: %v", err)
		return nil, nil
	}

	s.longTerm.SetCachedKeywords(keywords, 30*time.Second)

	summaries, err := s.longTerm.SearchMemories(ctx, chatID, keywords, 5)
	if err != nil {
		s.logger.Warnf("tars: search memories error: %v", err)
		return nil, nil
	}

	return summaries, nil
}

func (s *MemoryService) TrySummarizeAndSave(ctx context.Context, chatID string) error {
	if s.llmClient == nil {
		return nil
	}

	currentLen := s.shortTerm.Len()
	if currentLen < s.cfg.Memory.MaxHistory {
		return nil
	}

	olderThanTTL := s.shortTerm.GetOlderThan(time.Duration(s.cfg.Memory.MemoryTTL) * time.Minute)
	if len(olderThanTTL) < s.cfg.Memory.MaxHistory/2 {
		return nil
	}

	summary, err := s.longTerm.SummarizeConversation(ctx, olderThanTTL)
	if err != nil {
		return err
	}

	keywords, err := s.longTerm.ExtractKeywords(ctx, olderThanTTL)
	if err != nil {
		return err
	}

	keywordStr := strings.Join(keywords, ",")
	startTime := olderThanTTL[0].Time
	endTime := olderThanTTL[len(olderThanTTL)-1].Time

	return s.longTerm.CreateMemory(ctx, chatID, keywordStr, summary, startTime, endTime, len(olderThanTTL))
}

func (s *MemoryService) ClearHistory(chatID string) {
	s.shortTerm.Clear()
}

func (s *MemoryService) CleanupOld(ctx context.Context) error {
	return s.longTerm.CleanupOldData(ctx)
}
