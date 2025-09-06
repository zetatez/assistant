package tars

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/channel"
	"assistant/pkg/llm"
)

type Handler struct {
	ch        channel.Channel
	llmClient llm.Client
	memory    *MemoryService
	wikiRepo  *repo.WikiRepo
	logger    Logger
	llmModel  string
}

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
}

func NewHandler(ch channel.Channel, memory *MemoryService, wikiRepo *repo.WikiRepo, logger Logger) *Handler {
	llmCfg := psl.GetConfig().LLM

	var llmClient llm.Client
	if llmCfg.Provider != "" {
		var err error
		llmClient, err = llm.NewClient(llmCfg.Provider, llm.Config{
			APIKey:     llmCfg.APIKey,
			BaseURL:    llmCfg.BaseURL,
			Model:      llmCfg.Model,
			Timeout:    llmCfg.Timeout,
			MaxRetries: 3,
		})
		if err != nil {
			logger.Errorf("tars: failed to create LLM client: %v", err)
		}
	}

	return &Handler{
		ch:        ch,
		memory:    memory,
		wikiRepo:  wikiRepo,
		logger:    logger,
		llmClient: llmClient,
		llmModel:  llmCfg.Model,
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

	if event.MsgType != "text" {
		h.logger.Infof("tars: skipping non-text message type: %s", event.MsgType)
		return
	}

	var content struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(event.Content), &content); err != nil {
		h.logger.Errorf("tars: failed to parse message content: %v", err)
		return
	}

	text := strings.TrimSpace(content.Text)
	if text == "" {
		return
	}

	if strings.HasPrefix(text, "/clear") {
		h.memory.ClearHistory(event.ChatID)
		h.reply(event.ChatID, "对话历史已清除")
		return
	}

	go h.processMessage(event.ChatID, event.OpenID, event.MessageID, text)
}

func (h *Handler) processMessage(chatID, openID, messageID, userMessage string) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	username := openID
	if err := h.memory.AddUserMessage(ctx, chatID, openID, username, userMessage, messageID); err != nil {
		h.logger.Errorf("tars: save user message error: %v", err)
	}

	messages, err := h.memory.GetContextForLLM(ctx, chatID)
	if err != nil {
		h.logger.Errorf("tars: get context error: %v", err)
		h.reply(chatID, "抱歉，获取对话上下文失败")
		return
	}

	if h.llmClient == nil {
		h.logger.Errorf("tars: LLM client is nil, cannot process message")
		h.reply(chatID, "抱歉，AI服务未配置")
		return
	}

	wikiContext := h.searchWikiContext(ctx, userMessage)
	if wikiContext != "" {
		wikiMsg := llm.Message{
			Role:    llm.RoleSystem,
			Content: "【知识库参考】\n" + wikiContext + "\n\n请根据上述知识库内容回答用户问题。如果知识库中有相关信息，请结合知识库回答。",
		}
		messages = append([]llm.Message{wikiMsg}, messages...)
	}

	req := llm.ChatRequest{
		Model:       h.llmModel,
		Messages:    messages,
		Temperature: psl.GetConfig().Tars.LLMTemperature,
	}

	resp, err := h.llmClient.Chat(ctx, req)
	if err != nil {
		h.logger.Errorf("tars: LLM chat error: %v", err)
		h.reply(chatID, "抱歉，AI服务处理失败")
		return
	}

	reply := strings.TrimSpace(resp.Content)
	if reply != "" {
		if err := h.memory.AddAssistantMessage(ctx, chatID, openID, username, reply); err != nil {
			h.logger.Errorf("tars: save assistant message error: %v", err)
		}

		if err := h.memory.TrySummarizeAndSave(ctx, chatID); err != nil {
			h.logger.Errorf("tars: summarize error: %v", err)
		}

		h.reply(chatID, reply)
	}
}

func (h *Handler) searchWikiContext(ctx context.Context, query string) string {
	if h.wikiRepo == nil {
		return ""
	}

	entries, err := h.wikiRepo.Search(ctx, query, 3)
	if err != nil {
		h.logger.Errorf("tars: wiki search error: %v", err)
		return ""
	}

	if len(entries) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, entry := range entries {
		sb.WriteString("## ")
		sb.WriteString(entry.Title)
		sb.WriteString("\n")
		if len(entry.Content) > 2000 {
			sb.WriteString(entry.Content[:2000])
			sb.WriteString("...[内容过长已截断]")
		} else {
			sb.WriteString(entry.Content)
		}
		sb.WriteString("\n\n")
	}
	return sb.String()
}

func (h *Handler) reply(chatID, text string) {
	if h.ch == nil {
		return
	}

	content, _ := json.Marshal(map[string]string{"text": text})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.ch.SendMessage(ctx, chatID, "text", string(content)); err != nil {
		h.logger.Errorf("tars: failed to send reply: %v", err)
	}
}
