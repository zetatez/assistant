package tars

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"assistant/pkg/channel"
	"assistant/pkg/llmproxy"
)

const (
	CommandClear = "/clear"
	ReplyTimeout = 10 * time.Second

	maxContextTokens = 12000
)

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
}

type Handler struct {
	ch             channel.Channel
	llmClient      llmproxy.Client
	memory         *MemoryService
	logger         Logger
	llmModel       string
	llmTemperature float32
	messageTimeout time.Duration
	tools          *ToolExecutor
	msgParser      *Parser
	sessionLocks   sync.Map
}

func NewHandler(ch channel.Channel, memory *MemoryService, llmClient llmproxy.Client, wikiMgr *IndexManager, logger Logger, llmModel string, llmTemperature float32, messageTimeout time.Duration) *Handler {
	return &Handler{
		ch:             ch,
		memory:         memory,
		llmClient:      llmClient,
		logger:         logger,
		llmModel:       llmModel,
		llmTemperature: llmTemperature,
		messageTimeout: messageTimeout,
		tools:          NewToolExecutor(wikiMgr, logger),
		msgParser:      NewParser(logger),
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

	go h.processMessage(event.SessionID, event.OpenID, event.MessageID, userMsg)
}

func (h *Handler) lockSession(sessionID string) func() {
	v, _ := h.sessionLocks.LoadOrStore(sessionID, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

func (h *Handler) processMessage(sessionID, openID, messageID string, userMsg UserMessage) {
	unlock := h.lockSession(sessionID)
	defer unlock()

	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), h.messageTimeout)
	defer cancel()

	traceID := generateTraceID()
	h.logger.Infof("tars: [trace=%s] processing message, sessionID=%s, openID=%s", traceID, sessionID, openID)

	if h.llmClient == nil {
		h.logger.Errorf("tars: [trace=%s] LLM client is nil", traceID)
		h.reply(sessionID, "AI service not available, please try again later")
		return
	}

	userText := userMsg.Text
	if err := h.memory.AddUserMessage(ctx, sessionID, openID, openID, userText, messageID); err != nil {
		h.logger.Errorf("tars: [trace=%s] save user message error: %v", traceID, err)
	}

	messages, err := h.buildContext(ctx, sessionID)
	if err != nil {
		h.logger.Errorf("tars: [trace=%s] build context error: %v, using fallback", traceID, err)
		messages = h.buildFallbackContext(sessionID)
	}

	userLLMMsg := llmproxy.Message{
		Role:    llmproxy.RoleUser,
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

	reply, err := h.callLLM(ctx, messages)
	if err != nil {
		h.logger.Errorf("tars: [trace=%s] LLM error: %v", traceID, err)
		errMsg := err.Error()
		if len(errMsg) > 100 {
			errMsg = errMsg[:100] + "..."
		}
		h.reply(sessionID, fmt.Sprintf("AI processing failed: %s, please try again later", errMsg))
		return
	}

	reply = strings.TrimSpace(reply)
	if reply == "" {
		return
	}

	if err := h.memory.AddAssistantMessage(ctx, sessionID, openID, openID, reply); err != nil {
		h.logger.Errorf("tars: [trace=%s] save assistant message error: %v", traceID, err)
	}

	go func() {
		summCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := h.memory.TrySummarizeAndSave(summCtx, sessionID); err != nil {
			h.logger.Warnf("tars: summarize error: %v", err)
		}
	}()

	h.reply(sessionID, reply)

	h.logger.Infof("tars: [trace=%s] message processed in %v", traceID, time.Since(startTime))
}

func (h *Handler) buildContext(ctx context.Context, sessionID string) ([]llmproxy.Message, error) {
	memMsgs, err := h.memory.GetContextForLLM(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	recentMsgs, err := h.memory.GetRecentMessages(ctx, sessionID)
	if err != nil {
		return memMsgs, nil
	}
	messages := append(memMsgs, recentMsgs...)
	return truncateMessagesByTokens(messages, maxContextTokens), nil
}

func (h *Handler) buildFallbackContext(sessionID string) []llmproxy.Message {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	messages, err := h.memory.GetContextForLLM(ctx, sessionID)
	if err != nil || len(messages) == 0 {
		messages = []llmproxy.Message{
			{Role: llmproxy.RoleSystem, Content: loadSystemPrompt()},
		}
	}
	recentMsgs, err := h.memory.GetRecentMessages(ctx, sessionID)
	if err == nil && len(recentMsgs) > 0 {
		messages = append(messages, recentMsgs...)
	}
	return truncateMessagesByTokens(messages, maxContextTokens)
}

func (h *Handler) callLLM(ctx context.Context, messages []llmproxy.Message) (string, error) {
	req := llmproxy.ChatRequest{
		Model:       h.llmModel,
		Messages:    messages,
		Tools:       h.tools.Definitions(),
		Temperature: h.llmTemperature,
	}
	return runReAct(ctx, h.llmClient, req, h.tools, h.logger)
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
	rand.Read(b)
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}
