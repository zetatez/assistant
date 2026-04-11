package tars

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"assistant/internal/app/modules/tars/knowledge"
	"assistant/internal/app/modules/tars/message"
	"assistant/pkg/channel"
	"assistant/pkg/llm"
	"assistant/pkg/wiki"
	"assistant/pkg/workpool"
)

const (
	CommandClear   = "/clear"
	MessageTimeout = 120 * time.Second
	ReplyTimeout   = 10 * time.Second
	DBTimeout      = 3 * time.Second
)

type Metrics struct {
	mu                    sync.RWMutex
	messages              int64
	errors                int64
	processingMs          int64
	summarizeSuccess      int64
	summarizeFail         int64
	knowledgeSuccess      int64
	knowledgeFail         int64
	sessionRefreshSuccess int64
	sessionRefreshFail    int64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) IncMessage() {
	atomic.AddInt64(&m.messages, 1)
}

func (m *Metrics) IncError() {
	atomic.AddInt64(&m.errors, 1)
}

func (m *Metrics) RecordDuration(ms int64) {
	atomic.AddInt64(&m.processingMs, ms)
}

func (m *Metrics) IncSummarizeSuccess() {
	atomic.AddInt64(&m.summarizeSuccess, 1)
}

func (m *Metrics) IncSummarizeFail() {
	atomic.AddInt64(&m.summarizeFail, 1)
}

func (m *Metrics) IncKnowledgeSuccess() {
	atomic.AddInt64(&m.knowledgeSuccess, 1)
}

func (m *Metrics) IncKnowledgeFail() {
	atomic.AddInt64(&m.knowledgeFail, 1)
}

func (m *Metrics) IncSessionRefreshSuccess() {
	atomic.AddInt64(&m.sessionRefreshSuccess, 1)
}

func (m *Metrics) IncSessionRefreshFail() {
	atomic.AddInt64(&m.sessionRefreshFail, 1)
}

func (m *Metrics) Stats() (messages, errors, avgMs int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.messages, m.errors, m.processingMs / max(m.messages, 1)
}

func (m *Metrics) BGStats() (summarizeSuccess, summarizeFail, knowledgeSuccess, knowledgeFail, sessionRefreshSuccess, sessionRefreshFail int64) {
	return m.summarizeSuccess, m.summarizeFail, m.knowledgeSuccess, m.knowledgeFail, m.sessionRefreshSuccess, m.sessionRefreshFail
}

const maxConcurrentLLMTasks = 20
const knowledgeExtractInterval = 5

const minWikiRelevanceScore = 5.0

const maxContextTokens = 12000

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

	kept := make([]llm.Message, 0, len(messages))
	remaining := maxTokens
	for i := len(messages) - 1; i >= 0; i-- {
		m := messages[i]
		tokens := estimateTokens(m.Content)
		if tokens <= remaining {
			kept = append([]llm.Message{m}, kept...)
			remaining -= tokens
		}
		if remaining <= 0 {
			break
		}
	}
	return kept
}

// LLM semaphore priority: main request LLM calls (callLLM) bypass the semaphore
// entirely, while background tasks (summarize, knowledge extract, session refresh)
// acquire a slot first. This ensures main requests never wait even when background
// tasks saturate the LLM capacity.

type Handler struct {
	ch                  channel.Channel
	llmClient           llm.Client
	memory              IMemoryService
	logger              Logger
	llmModel            string
	llmTemperature      float32
	pool                *workpool.WorkPool
	metrics             *Metrics
	knowledgeManager    *knowledge.Manager
	sessionManager      *knowledge.SessionManager
	wikiManager         *wiki.IndexManager
	llmSem              chan struct{}
	msgCounter          map[string]int
	msgCounterMu        sync.RWMutex
	sessionMsgCounter   map[string]int
	sessionMsgCounterMu sync.RWMutex
	msgParser           *message.Parser
}

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
}

const maxQueueSize = 10000

func NewHandler(ch channel.Channel, memory IMemoryService, llmClient llm.Client, knowledgeMgr *knowledge.Manager, sessionMgr *knowledge.SessionManager, wikiMgr *wiki.IndexManager, logger Logger, llmModel string, llmTemperature float32) *Handler {
	pool := workpool.New("tars", 50, maxQueueSize)
	pool.Start(10)

	return &Handler{
		ch:                ch,
		memory:            memory,
		logger:            logger,
		llmClient:         llmClient,
		llmModel:          llmModel,
		llmTemperature:    llmTemperature,
		pool:              pool,
		metrics:           NewMetrics(),
		knowledgeManager:  knowledgeMgr,
		sessionManager:    sessionMgr,
		wikiManager:       wikiMgr,
		llmSem:            make(chan struct{}, maxConcurrentLLMTasks),
		msgCounter:        make(map[string]int),
		sessionMsgCounter: make(map[string]int),
		msgParser:         message.NewParser(logger),
	}
}

func (h *Handler) Register() {
	if h.ch == nil {
		return
	}
	h.ch.SetMessageHandler(h)
}

func (h *Handler) OnMessageReceive(event *channel.MessageEvent) {
	if event == nil {
		return
	}

	userMsg := h.msgParser.Parse(event.MsgType, event.Content)
	if !userMsg.Supported {
		return
	}
	if userMsg.Skip {
		return
	}
	if userMsg.Text == "" && userMsg.ImageKey == "" && userMsg.FileKey == "" {
		return
	}

	if strings.HasPrefix(userMsg.Text, CommandClear) {
		if err := h.memory.ClearHistory(event.SessionID); err != nil {
			h.logger.Errorf("tars: clear history error: %v", err)
		}
		h.reply(event.SessionID, "Conversation history cleared")
		return
	}

	ok := h.pool.Go(func() {
		h.processMessage(event.SessionID, event.OpenID, event.MessageID, userMsg)
	})
	if !ok {
		h.logger.Warnf("tars: message dropped, pool full, sessionID=%s", event.SessionID)
		h.reply(event.SessionID, "System is currently busy, please try again in a moment.")
	}
}

func (h *Handler) processMessage(sessionID, openID, messageID string, userMsg message.UserMessage) {
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), MessageTimeout)
	defer cancel()

	traceID := generateTraceID()
	ctx = context.WithValue(ctx, "traceID", traceID)

	h.logger.Infof("tars: [trace=%s] processing message, sessionID=%s, openID=%s", traceID, sessionID, openID)
	h.metrics.IncMessage()

	userText := userMsg.Text
	if err := h.saveUserMessage(ctx, sessionID, openID, userText, messageID); err != nil {
		h.logger.Errorf("tars: [trace=%s] save user message error: %v", traceID, err)
	}

	messages, err := h.buildContext(ctx, sessionID, userText)
	if err != nil {
		h.logger.Errorf("tars: [trace=%s] build context error: %v, using fallback", traceID, err)
		messages, _ = h.buildFallbackContext(sessionID)
		if len(messages) == 0 {
			h.logger.Errorf("tars: [trace=%s] fallback context also failed", traceID)
		}
	}

	if h.llmClient == nil {
		h.logger.Errorf("tars: [trace=%s] LLM client is nil", traceID)
		h.reply(sessionID, "AI service not available, please try again later")
		h.metrics.IncError()
		return
	}

	recallCtx, recallCancel := context.WithTimeout(ctx, 3*time.Second)
	defer recallCancel()
	recallContent, err := h.memory.TryRecallIfNeeded(recallCtx, sessionID, userText)
	if err == nil && recallContent != "" {
		recallMsg := llm.Message{
			Role:    llm.RoleSystem,
			Content: recallContent,
		}
		messages = append(messages, recallMsg)
	}

	userLLMMsg := llm.Message{
		Role:    llm.RoleUser,
		Content: userText,
	}

	var imgData []byte
	var fileData []byte
	var imgErr, fileErr error

	var dlWg sync.WaitGroup
	if userMsg.ImageKey != "" {
		dlWg.Add(1)
		go func() {
			defer dlWg.Done()
			imgData, _, imgErr = h.ch.DownloadMedia(ctx, messageID, userMsg.ImageKey)
		}()
	}
	if userMsg.FileKey != "" && userMsg.FileName != "" {
		dlWg.Add(1)
		go func() {
			defer dlWg.Done()
			fileData, _, fileErr = h.ch.DownloadMedia(ctx, messageID, userMsg.FileKey)
		}()
	}
	dlWg.Wait()

	if userMsg.ImageKey != "" && imgErr == nil && len(imgData) > 0 {
		userLLMMsg.ImageBase64 = base64.StdEncoding.EncodeToString(imgData)
		h.logger.Infof("tars: [trace=%s] image downloaded, size=%d", traceID, len(imgData))
	} else if userMsg.ImageKey != "" && imgErr != nil {
		h.logger.Warnf("tars: [trace=%s] download image error: %v", traceID, imgErr)
	}

	if userMsg.FileKey != "" && userMsg.FileName != "" && fileErr == nil && len(fileData) > 0 {
		fileContent := string(fileData)
		if len(fileContent) > 10000 {
			fileContent = fileContent[:10000] + "\n... [truncated]"
		}
		userLLMMsg.Content += fmt.Sprintf("\n\n[File: %s content]:\n%s", userMsg.FileName, fileContent)
		h.logger.Infof("tars: [trace=%s] text file downloaded, size=%d", traceID, len(fileData))
	} else if userMsg.FileKey != "" && userMsg.FileName != "" && fileErr != nil {
		h.logger.Warnf("tars: [trace=%s] download file error: %v", traceID, fileErr)
	}

	messages = append(messages, userLLMMsg)

	resp, err := h.callLLM(ctx, messages)
	if err != nil {
		h.logger.Errorf("tars: [trace=%s] LLM error: %v", traceID, err)
		errMsg := err.Error()
		if len(errMsg) > 100 {
			errMsg = errMsg[:100] + "..."
		}
		h.reply(sessionID, fmt.Sprintf("AI processing failed: %s, please try again later", errMsg))
		h.metrics.IncError()
		return
	}

	if err := h.handleResponse(ctx, sessionID, openID, userText, resp); err != nil {
		h.logger.Errorf("tars: [trace=%s] handle response error: %v", traceID, err)
	}

	h.metrics.RecordDuration(time.Since(startTime).Milliseconds())
	h.logger.Infof("tars: [trace=%s] message processed in %v", traceID, time.Since(startTime))
}

func (h *Handler) saveUserMessage(ctx context.Context, sessionID, openID, content, messageID string) error {
	dbCtx, cancel := context.WithTimeout(ctx, DBTimeout)
	defer cancel()
	return h.memory.AddUserMessage(dbCtx, sessionID, openID, openID, content, messageID)
}

func (h *Handler) buildContext(ctx context.Context, sessionID string, userQuery string) ([]llm.Message, error) {
	type ctxResult struct {
		msgs []llm.Message
		err  error
	}

	var wg sync.WaitGroup
	wg.Add(4)

	var memResult, sessResult, kgResult, wikiResult ctxResult

	go func() {
		defer wg.Done()
		memResult.msgs, memResult.err = h.memory.GetContextForLLM(ctx, sessionID)
	}()

	go func() {
		defer wg.Done()
		if h.sessionManager != nil {
			sessCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			state, err := h.sessionManager.GetSession(sessCtx, sessionID)
			if err == nil && (state.Summary != "" || state.Context != "" || len(state.PendingTasks) > 0) {
				var sb strings.Builder
				sb.WriteString("## Session State (only use if relevant to current question)\n\n")
				if state.Summary != "" {
					sb.WriteString("**Summary**: " + state.Summary + "\n\n")
				}
				if state.Context != "" {
					sb.WriteString("**Context**: " + state.Context + "\n\n")
				}
				if len(state.PendingTasks) > 0 {
					sb.WriteString("**Pending Tasks**:\n")
					for _, task := range state.PendingTasks {
						sb.WriteString("- " + task.Content + " [" + task.Status + "]\n")
					}
				}
				sessResult.msgs = append(sessResult.msgs, llm.Message{Role: llm.RoleSystem, Content: sb.String()})
			}
		}
	}()

	go func() {
		defer wg.Done()
		if h.knowledgeManager != nil && userQuery != "" {
			kgCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			knowledgeContext, err := h.knowledgeManager.GetRelatedKnowledge(kgCtx, sessionID, userQuery)
			if err == nil && knowledgeContext != "" {
				kgResult.msgs = append(kgResult.msgs, llm.Message{Role: llm.RoleSystem, Content: "## Related Knowledge (only use if relevant to current question)\n\n" + knowledgeContext})
			}
		}
	}()

	go func() {
		defer wg.Done()
		if h.wikiManager != nil && userQuery != "" {
			wikiCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			rerankResults, err := h.wikiManager.Search(wikiCtx, userQuery, 10)
			if err == nil && len(rerankResults) > 0 {
				var sb strings.Builder
				sb.WriteString("## Local Wiki Knowledge (only use if directly relevant to current question)\n\n")
				count := 0
				for _, res := range rerankResults {
					if res.Score < minWikiRelevanceScore {
						continue
					}
					if count >= 3 {
						break
					}
					count++
					title := res.Entry.Title
					if title == "" {
						title = filepath.Base(res.Entry.Path)
					}
					sb.WriteString("### " + title + " (relevance: " + fmt.Sprintf("%.1f", res.Score) + ")\n\n")
					sb.WriteString(res.Snippet)
					sb.WriteString("\n\n")
				}
				if count > 0 {
					wikiResult.msgs = append(wikiResult.msgs, llm.Message{Role: llm.RoleSystem, Content: sb.String()})
				}
			}
		}
	}()

	wg.Wait()

	if memResult.err != nil {
		return memResult.msgs, memResult.err
	}

	var recentMsgs []llm.Message
	var recentErr error
	var recentWg sync.WaitGroup
	recentWg.Add(1)
	go func() {
		defer recentWg.Done()
		recentMsgs, recentErr = h.memory.GetRecentMessages(ctx, sessionID)
	}()
	recentWg.Wait()

	var messages []llm.Message
	messages = append(messages, memResult.msgs...)
	messages = append(messages, sessResult.msgs...)
	messages = append(messages, wikiResult.msgs...)
	messages = append(messages, kgResult.msgs...)
	if recentErr == nil && len(recentMsgs) > 0 {
		messages = append(messages, recentMsgs...)
	}

	messages = truncateMessagesByTokens(messages, maxContextTokens)

	if len(messages) > 0 && messages[len(messages)-1].Role == llm.RoleUser && messages[len(messages)-1].Content == userQuery {
		messages = messages[:len(messages)-1]
	}

	return messages, nil
}

func (h *Handler) buildFallbackContext(sessionID string) ([]llm.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	messages, err := h.memory.GetContextForLLM(ctx, sessionID)
	if err != nil || len(messages) == 0 {
		messages = []llm.Message{
			{Role: llm.RoleSystem, Content: getSystemPrompt(50, 80, h.llmModel)},
		}
	}
	recentMsgs, recentErr := h.memory.GetRecentMessages(ctx, sessionID)
	if recentErr != nil {
		h.logger.Warnf("tars: buildFallbackContext: get recent messages error: %v", recentErr)
	} else if len(recentMsgs) > 0 {
		messages = append(messages, recentMsgs...)
	}
	return truncateMessagesByTokens(messages, maxContextTokens), nil
}

func (h *Handler) callLLM(ctx context.Context, messages []llm.Message) (*llm.ChatResponse, error) {
	req := llm.ChatRequest{
		Model:       h.llmModel,
		Messages:    messages,
		Temperature: h.llmTemperature,
	}
	return h.llmClient.Chat(ctx, req)
}

func (h *Handler) handleResponse(ctx context.Context, sessionID, openID, userMsg string, resp *llm.ChatResponse) error {
	reply := strings.TrimSpace(resp.Content)
	if reply == "" {
		return nil
	}

	dbCtx, cancel := context.WithTimeout(ctx, DBTimeout)
	defer cancel()
	if err := h.memory.AddAssistantMessage(dbCtx, sessionID, openID, openID, reply); err != nil {
		h.logger.Errorf("tars: save assistant message error: %v", err)
	}

	if h.knowledgeManager != nil {
		h.msgCounterMu.Lock()
		if h.msgCounter == nil {
			h.msgCounter = make(map[string]int)
		}
		h.msgCounter[sessionID]++
		shouldExtract := h.msgCounter[sessionID] >= knowledgeExtractInterval
		if shouldExtract {
			h.msgCounter[sessionID] = 0
		}
		h.msgCounterMu.Unlock()

		if shouldExtract {
			h.llmSem <- struct{}{}
			kgCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
			go func() {
				defer func() {
					cancel()
					<-h.llmSem
				}()
				if err := h.knowledgeManager.IntegrateMessage(kgCtx, sessionID, userMsg, reply); err != nil {
					h.logger.Warnf("tars: knowledge integrate error: %v", err)
					h.metrics.IncKnowledgeFail()
				} else {
					h.metrics.IncKnowledgeSuccess()
				}
			}()
		}
	}

	if h.sessionManager != nil {
		h.sessionMsgCounterMu.Lock()
		if h.sessionMsgCounter == nil {
			h.sessionMsgCounter = make(map[string]int)
		}
		h.sessionMsgCounter[sessionID]++
		shouldRefreshSession := h.sessionMsgCounter[sessionID] >= knowledgeExtractInterval
		if shouldRefreshSession {
			h.sessionMsgCounter[sessionID] = 0
		}
		h.sessionMsgCounterMu.Unlock()

		if shouldRefreshSession {
			h.llmSem <- struct{}{}
			sessCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			go func() {
				defer func() {
					cancel()
					<-h.llmSem
				}()
				if err := h.sessionManager.RefreshSessionSummary(sessCtx, sessionID, []string{userMsg, reply}); err != nil {
					h.logger.Warnf("tars: session refresh error: %v", err)
					h.metrics.IncSessionRefreshFail()
				} else {
					h.metrics.IncSessionRefreshSuccess()
				}
			}()
		}
	}

	h.llmSem <- struct{}{}
	summCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	go func() {
		defer func() {
			cancel()
			<-h.llmSem
		}()
		if err := h.memory.TrySummarizeAndSave(summCtx, sessionID); err != nil {
			h.logger.Warnf("tars: summarize error: %v", err)
			h.metrics.IncSummarizeFail()
		} else {
			h.metrics.IncSummarizeSuccess()
		}
	}()

	return h.reply(sessionID, reply)
}

func (h *Handler) reply(sessionID, text string) error {
	if h.ch == nil {
		return nil
	}

	content, _ := json.Marshal(map[string]string{"text": text})
	ctx, cancel := context.WithTimeout(context.Background(), ReplyTimeout)
	defer cancel()

	var lastErr error
	for i := 0; i < 3; i++ {
		if err := h.ch.SendMessage(ctx, sessionID, "text", string(content)); err == nil {
			return nil
		} else {
			lastErr = err
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}
	}
	h.logger.Errorf("tars: failed to send reply after 3 retries: %v", lastErr)
	return lastErr
}

func generateTraceID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

func (h *Handler) Stop() {
	if h.pool != nil {
		h.pool.Stop()
	}
}
