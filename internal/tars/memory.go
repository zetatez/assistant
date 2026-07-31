package tars

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	llm "assistant/pkg/llmproxy"
)

var (
	defaultMaxHistory  = 64
	summarizeThreshold = 10
	maxHistoryBytes    = int64(64 * 1024 * 1024)
)

var memoryDocTemplate = `# Memory Document

## User Profile
(none)

## Preferences
(none)

## Key Facts
(none)

## Important Decisions
(none)

## In Progress
(none)

## Pending
(none)

## Completed
(none)

## Recent Context
(none)
`

type ShortTermMessage struct {
	Role    string
	Content string
	Time    time.Time
}

type chatBuffer struct {
	messages   []ShortTermMessage
	head       int
	count      int
	mu         sync.Mutex
	lastAccess time.Time
}

type shortTerm struct {
	capacity int
	buffers  sync.Map
	initMu   sync.Mutex
}

func newShortTerm(capacity int) *shortTerm {
	if capacity <= 0 {
		capacity = defaultMaxHistory
	}
	return &shortTerm{capacity: capacity}
}

func (s *shortTerm) getOrCreateBuffer(sessionID string) *chatBuffer {
	if v, ok := s.buffers.Load(sessionID); ok {
		return v.(*chatBuffer)
	}
	s.initMu.Lock()
	if v, ok := s.buffers.Load(sessionID); ok {
		s.initMu.Unlock()
		return v.(*chatBuffer)
	}
	buf := &chatBuffer{
		messages:   make([]ShortTermMessage, s.capacity),
		lastAccess: time.Now(),
	}
	s.buffers.Store(sessionID, buf)
	s.initMu.Unlock()
	return buf
}

func (s *shortTerm) Add(sessionID, role, content string) {
	buf := s.getOrCreateBuffer(sessionID)
	buf.mu.Lock()
	defer buf.mu.Unlock()
	buf.lastAccess = time.Now()
	buf.messages[buf.head] = ShortTermMessage{
		Role:    role,
		Content: content,
		Time:    time.Now(),
	}
	buf.head = (buf.head + 1) % s.capacity
	if buf.count < s.capacity {
		buf.count++
	}
}

func (s *shortTerm) GetAll(sessionID string) []ShortTermMessage {
	v, ok := s.buffers.Load(sessionID)
	if !ok {
		return nil
	}
	buf := v.(*chatBuffer)
	buf.mu.Lock()
	defer buf.mu.Unlock()
	buf.lastAccess = time.Now()
	if buf.count == 0 {
		return nil
	}
	result := make([]ShortTermMessage, buf.count)
	for i := 0; i < buf.count; i++ {
		idx := (buf.head - buf.count + i + s.capacity) % s.capacity
		result[i] = buf.messages[idx]
	}
	return result
}

func (s *shortTerm) Clear(sessionID string) {
	v, ok := s.buffers.Load(sessionID)
	if !ok {
		return
	}
	buf := v.(*chatBuffer)
	buf.mu.Lock()
	defer buf.mu.Unlock()
	buf.head = 0
	buf.count = 0
}

func (s *shortTerm) CleanupOldSessions(maxAge time.Duration) int {
	cutoff := time.Now().Add(-maxAge)
	var toDelete []string
	s.buffers.Range(func(key, value any) bool {
		buf := value.(*chatBuffer)
		buf.mu.Lock()
		if buf.lastAccess.Before(cutoff) {
			toDelete = append(toDelete, key.(string))
		}
		buf.mu.Unlock()
		return true
	})
	for _, id := range toDelete {
		s.buffers.Delete(id)
	}
	return len(toDelete)
}

type MemoryService struct {
	shortTerm   *shortTerm
	dataDir     string
	llmClient   llm.Client
	llmModel    string
	logger      Logger
	msgCounters sync.Map
	loaded      sync.Map
}

func NewMemoryService(dataDir string, llmClient llm.Client, llmModel string, logger Logger) *MemoryService {
	dir := filepath.Join(dataDir, "tars")
	os.MkdirAll(dir, 0755)
	return &MemoryService{
		shortTerm: newShortTerm(defaultMaxHistory),
		dataDir:   dir,
		llmClient: llmClient,
		llmModel:  llmModel,
		logger:    logger,
	}
}

func (m *MemoryService) AddUserMessage(ctx context.Context, chatID, openID, username, content, messageID string) error {
	m.ensureLoaded(chatID)
	m.shortTerm.Add(chatID, "user", content)
	m.appendHistory(chatID, "user", content)
	m.incMsgCount(chatID)
	return nil
}

func (m *MemoryService) AddAssistantMessage(ctx context.Context, chatID, openID, username, content string) error {
	m.ensureLoaded(chatID)
	m.shortTerm.Add(chatID, "assistant", content)
	m.appendHistory(chatID, "assistant", content)
	m.incMsgCount(chatID)
	return nil
}

func (m *MemoryService) GetContextForLLM(ctx context.Context, chatID string) ([]llm.Message, error) {
	m.ensureLoaded(chatID)
	var messages []llm.Message
	messages = append(messages, llm.Message{
		Role:    llm.RoleSystem,
		Content: loadSystemPrompt(),
	})
	memDoc := m.loadMemoryDoc(chatID)
	if memDoc != "" {
		messages = append(messages, llm.Message{
			Role:    llm.RoleSystem,
			Content: "## Long-term Memory (only use if relevant to current question)\n\n" + memDoc,
		})
	}
	return messages, nil
}

func (m *MemoryService) GetRecentMessages(ctx context.Context, chatID string) ([]llm.Message, error) {
	m.ensureLoaded(chatID)
	msgs := m.shortTerm.GetAll(chatID)
	return toLLMMessages(msgs), nil
}

func (m *MemoryService) ClearHistory(chatID string) error {
	m.shortTerm.Clear(chatID)
	m.loaded.Delete(chatID)
	m.msgCounters.Delete(chatID)
	dir := m.sessionDir(chatID)
	os.RemoveAll(dir)
	return nil
}

func (m *MemoryService) TrySummarizeAndSave(ctx context.Context, chatID string) error {
	if m.llmClient == nil {
		return nil
	}
	v, ok := m.msgCounters.LoadAndDelete(chatID)
	if !ok {
		return nil
	}
	count := v.(int)
	if count < summarizeThreshold {
		m.msgCounters.Store(chatID, count)
		return nil
	}

	msgs := m.shortTerm.GetAll(chatID)
	if len(msgs) == 0 {
		return nil
	}

	_, err := m.updateMemoryDoc(ctx, chatID, msgs)
	if err != nil {
		m.logger.Warnf("memory: summarize error: %v", err)
		m.msgCounters.Store(chatID, count)
		return err
	}

	return nil
}

func (m *MemoryService) updateMemoryDoc(ctx context.Context, chatID string, msgs []ShortTermMessage) (string, error) {
	existing := m.loadMemoryDoc(chatID)
	if existing == "" {
		existing = memoryDocTemplate
	}

	var convo strings.Builder
	for _, msg := range msgs {
		convo.WriteString(msg.Role)
		convo.WriteString(": ")
		convo.WriteString(msg.Content)
		convo.WriteString("\n")
	}

	prompt := fmt.Sprintf(`You are maintaining a memory document for an ongoing conversation. Update the memory document based on the recent conversation.

Rules:
- Preserve existing information that is still relevant.
- Add new information from the recent conversation to the appropriate sections.
- **User Profile**: name, role, timezone, language, communication style.
- **Preferences**: answer format, tech stack preferences, topics of interest.
- **Key Facts**: project context, tech stack, environment info, important dates, domain knowledge.
- **Important Decisions**: key decisions made and their rationale.
- **In Progress**: tasks and topics actively being worked on.
- **Pending**: todo items, unresolved questions, things waiting on.
- **Completed**: move finished items here from In Progress/Pending. Keep at most 3-5 recent items, discard older ones.
- **Recent Context**: 1-2 sentence summary of the latest conversation for continuity.
- Keep each section concise. Use bullet points. Do not invent information not present in the conversation.
- Output the COMPLETE updated memory document in Markdown, preserving all section headers exactly.
- Do not output anything other than the memory document.

## Current Memory Document
%s

## Recent Conversation
%s`, existing, convo.String())

	resp, err := m.llmClient.Chat(ctx, llm.ChatRequest{
		Model:       m.llmModel,
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: prompt}},
		Temperature: 0.3,
		MaxTokens:   800,
	})
	if err != nil {
		return "", err
	}

	updated := strings.TrimSpace(resp.Content)
	if updated == "" {
		return existing, nil
	}

	path := m.memoryPath(chatID)
	os.MkdirAll(m.sessionDir(chatID), 0755)
	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return "", err
	}

	return updated, nil
}

func (m *MemoryService) sessionDir(chatID string) string {
	return filepath.Join(m.dataDir, chatID)
}

func (m *MemoryService) historyPath(chatID string) string {
	return filepath.Join(m.sessionDir(chatID), "history.md")
}

func (m *MemoryService) memoryPath(chatID string) string {
	return filepath.Join(m.sessionDir(chatID), "memory.md")
}

func (m *MemoryService) appendHistory(chatID, role, content string) {
	dir := m.sessionDir(chatID)
	os.MkdirAll(dir, 0755)
	path := m.historyPath(chatID)
	line := fmt.Sprintf("### %s | %s\n%s\n\n", role, time.Now().Format(time.RFC3339), content)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		m.logger.Errorf("memory: failed to open history file: %v", err)
		return
	}
	defer f.Close()
	f.WriteString(line)

	if fi, err := f.Stat(); err == nil && fi.Size() > maxHistoryBytes {
		m.trimHistory(path, maxHistoryBytes/2)
	}
}

func (m *MemoryService) trimHistory(path string, targetBytes int64) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	if int64(len(data)) <= targetBytes {
		return
	}

	cut := len(data) - int(targetBytes)
	idx := strings.IndexByte(string(data[cut:]), '\n')
	if idx == -1 {
		return
	}
	start := cut + idx + 1

	sectionStart := strings.Index(string(data[start:]), "### ")
	if sectionStart == -1 {
		return
	}
	start += sectionStart

	os.WriteFile(path, data[start:], 0644)
}

func (m *MemoryService) loadHistory(chatID string) {
	path := m.historyPath(chatID)
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	text := string(data)
	sections := strings.Split(text, "\n### ")

	var parsed []ShortTermMessage
	for i, section := range sections {
		if i == 0 {
			if strings.HasPrefix(section, "### ") {
				if msg, ok := parseHistorySection(section); ok {
					parsed = append(parsed, msg)
				}
			}
			continue
		}
		if msg, ok := parseHistorySection("### " + section); ok {
			parsed = append(parsed, msg)
		}
	}

	start := 0
	if len(parsed) > defaultMaxHistory {
		start = len(parsed) - defaultMaxHistory
	}
	for _, msg := range parsed[start:] {
		m.shortTerm.Add(chatID, msg.Role, msg.Content)
	}
}

func parseHistorySection(section string) (ShortTermMessage, bool) {
	section = strings.TrimSpace(section)
	if section == "" {
		return ShortTermMessage{}, false
	}
	lines := strings.SplitN(section, "\n", 2)
	if len(lines) < 2 {
		return ShortTermMessage{}, false
	}
	header := strings.TrimPrefix(lines[0], "### ")
	parts := strings.SplitN(header, " | ", 2)
	if len(parts) < 1 {
		return ShortTermMessage{}, false
	}
	role := strings.TrimSpace(parts[0])
	content := strings.TrimSpace(lines[1])
	if role != "user" && role != "assistant" {
		return ShortTermMessage{}, false
	}
	return ShortTermMessage{Role: role, Content: content}, true
}

func (m *MemoryService) ensureLoaded(chatID string) {
	if _, loaded := m.loaded.LoadOrStore(chatID, true); loaded {
		return
	}
	m.loadHistory(chatID)
}

func (m *MemoryService) loadMemoryDoc(chatID string) string {
	path := m.memoryPath(chatID)
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func (m *MemoryService) incMsgCount(chatID string) {
	for {
		v, loaded := m.msgCounters.LoadOrStore(chatID, 1)
		if !loaded {
			return
		}
		if m.msgCounters.CompareAndSwap(chatID, v, v.(int)+1) {
			return
		}
	}
}

func toLLMMessages(msgs []ShortTermMessage) []llm.Message {
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

func estimateTokens(text string) int {
	return len(text) / 4
}

func truncateMessagesByTokens(messages []llm.Message, maxTokens int) []llm.Message {
	if len(messages) == 0 {
		return messages
	}
	total := 0
	for _, m := range messages {
		total += estimateTokens(m.Content)
	}
	if total <= maxTokens {
		return messages
	}

	var systemMsgs []llm.Message
	var rest []llm.Message
	for _, m := range messages {
		if m.Role == llm.RoleSystem {
			systemMsgs = append(systemMsgs, m)
		} else {
			rest = append(rest, m)
		}
	}

	sysTokens := 0
	for _, m := range systemMsgs {
		sysTokens += estimateTokens(m.Content)
	}

	remaining := maxTokens - sysTokens
	kept := make([]llm.Message, 0, len(messages))
	for i := len(rest) - 1; i >= 0; i-- {
		tokens := estimateTokens(rest[i].Content)
		if tokens <= remaining {
			kept = append([]llm.Message{rest[i]}, kept...)
			remaining -= tokens
		}
		if remaining <= 0 {
			break
		}
	}
	return append(systemMsgs, kept...)
}
