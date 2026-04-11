package chat

import (
	"strconv"

	"assistant/internal/app/repo"
	"assistant/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo   *repo.Queries
	logger Logger
}

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
}

func NewHandler(repo *repo.Queries, logger Logger) *Handler {
	return &Handler{
		repo:   repo,
		logger: logger,
	}
}

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/memory-doc/:session_id", h.getMemoryDoc)
	r.GET("/messages/:session_id", h.getMessages)
	r.GET("/messages/:session_id/search", h.searchMessages)
	r.GET("/memories/:session_id", h.getMemories)
	r.GET("/session/:session_id", h.getSession)
	r.GET("/entities/:session_id", h.getEntities)
	r.GET("/knowledge/:session_id", h.getKnowledge)
	r.GET("/recalls/:session_id", h.getRecalls)
}

// getMemoryDoc godoc
// @Summary 获取记忆文档
// @Description 获取指定会话的 Markdown 格式记忆文档
// @Tags Chat
// @Produce json
// @Param session_id path string true "会话ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /chat/memory-doc/{session_id} [get]
// @Security BearerAuth
func (h *Handler) getMemoryDoc(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		response.Err(c, response.CodeInvalidParams, "session_id is required")
		return
	}

	doc, err := h.repo.GetChatMemoryDoc(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Errorf("[chat] get memory doc failed: %v", err)
		response.ErrWithInternal(c, response.CodeDatabaseError, "failed to get memory doc", err)
		return
	}

	response.Ok(c, gin.H{
		"session_id": doc.SessionID,
		"content":    doc.Content,
		"version":    doc.Version,
		"updated_at": doc.UpdatedAt,
	})
}

// getMessages godoc
// @Summary 获取对话消息
// @Description 分页获取指定会话的对话消息列表
// @Tags Chat
// @Produce json
// @Param session_id path string true "会话ID"
// @Param offset query int false "偏移量" default(0)
// @Param limit query int false "返回数量" default(50)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /chat/messages/{session_id} [get]
// @Security BearerAuth
func (h *Handler) getMessages(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		response.Err(c, response.CodeInvalidParams, "session_id is required")
		return
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	messages, err := h.repo.GetChatMessages(c.Request.Context(), repo.GetChatMessagesParams{
		SessionID: sessionID,
		Offset:    int32(offset),
		Limit:     int32(limit),
	})
	if err != nil {
		h.logger.Errorf("[chat] get messages failed: %v", err)
		response.ErrWithInternal(c, response.CodeDatabaseError, "failed to get messages", err)
		return
	}

	response.Ok(c, gin.H{"messages": messages})
}

// searchMessages godoc
// @Summary 搜索对话消息
// @Description 根据关键词搜索指定会话的历史消息
// @Tags Chat
// @Produce json
// @Param session_id path string true "会话ID"
// @Param q query string true "搜索关键词"
// @Param limit query int false "返回数量" default(20)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /chat/messages/{session_id}/search [get]
// @Security BearerAuth
func (h *Handler) searchMessages(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		response.Err(c, response.CodeInvalidParams, "session_id is required")
		return
	}

	query := c.Query("q")
	if query == "" {
		response.Err(c, response.CodeInvalidParams, "query parameter 'q' is required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	pattern := "%" + query + "%"

	messages, err := h.repo.SearchChatMessagesByKeyword(c.Request.Context(), repo.SearchChatMessagesByKeywordParams{
		SessionID: sessionID,
		Content:   pattern,
		Limit:     int32(limit),
	})
	if err != nil {
		h.logger.Errorf("[chat] search messages failed: %v", err)
		response.ErrWithInternal(c, response.CodeDatabaseError, "failed to search messages", err)
		return
	}

	response.Ok(c, gin.H{"messages": messages})
}

// getMemories godoc
// @Summary 获取长期记忆摘要
// @Description 获取指定会话的长期记忆摘要列表
// @Tags Chat
// @Produce json
// @Param session_id path string true "会话ID"
// @Param limit query int false "返回数量" default(10)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /chat/memories/{session_id} [get]
// @Security BearerAuth
func (h *Handler) getMemories(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		response.Err(c, response.CodeInvalidParams, "session_id is required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	memories, err := h.repo.GetChatMemories(c.Request.Context(), repo.GetChatMemoriesParams{
		SessionID: sessionID,
		Limit:     int32(limit),
	})
	if err != nil {
		h.logger.Errorf("[chat] get memories failed: %v", err)
		response.ErrWithInternal(c, response.CodeDatabaseError, "failed to get memories", err)
		return
	}

	response.Ok(c, gin.H{"memories": memories})
}

// getSession godoc
// @Summary 获取会话状态
// @Description 获取指定会话的实时状态（摘要、待办事项、上下文）
// @Tags Chat
// @Produce json
// @Param session_id path string true "会话ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /chat/session/{session_id} [get]
// @Security BearerAuth
func (h *Handler) getSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		response.Err(c, response.CodeInvalidParams, "session_id is required")
		return
	}

	sess, err := h.repo.GetOrCreateChatSession(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Errorf("[chat] get session failed: %v", err)
		response.ErrWithInternal(c, response.CodeDatabaseError, "failed to get session", err)
		return
	}

	response.Ok(c, gin.H{
		"session_id":    sess.SessionID,
		"summary":       sess.Summary,
		"pending_tasks": sess.PendingTasks,
		"context":       sess.Context,
		"updated_at":    sess.UpdatedAt,
	})
}

// getEntities godoc
// @Summary 获取知识图谱实体
// @Description 分页获取指定会话的知识图谱实体列表
// @Tags Chat
// @Produce json
// @Param session_id path string true "会话ID"
// @Param limit query int false "返回数量" default(50)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /chat/entities/{session_id} [get]
// @Security BearerAuth
func (h *Handler) getEntities(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		response.Err(c, response.CodeInvalidParams, "session_id is required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	entities, err := h.repo.GetChatEntities(c.Request.Context(), repo.GetChatEntitiesParams{
		SessionID: sessionID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		h.logger.Errorf("[chat] get entities failed: %v", err)
		response.ErrWithInternal(c, response.CodeDatabaseError, "failed to get entities", err)
		return
	}

	response.Ok(c, gin.H{"entities": entities})
}

// getKnowledge godoc
// @Summary 获取知识页面
// @Description 分页获取指定会话的知识页面列表
// @Tags Chat
// @Produce json
// @Param session_id path string true "会话ID"
// @Param limit query int false "返回数量" default(50)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /chat/knowledge/{session_id} [get]
// @Security BearerAuth
func (h *Handler) getKnowledge(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		response.Err(c, response.CodeInvalidParams, "session_id is required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	knowledge, err := h.repo.GetChatKnowledge(c.Request.Context(), repo.GetChatKnowledgeParams{
		SessionID: sessionID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		h.logger.Errorf("[chat] get knowledge failed: %v", err)
		response.ErrWithInternal(c, response.CodeDatabaseError, "failed to get knowledge", err)
		return
	}

	response.Ok(c, gin.H{"knowledge": knowledge})
}

// getRecalls godoc
// @Summary 搜索历史召回记录
// @Description 根据关键词搜索指定会话的 LLM 召回历史记录
// @Tags Chat
// @Produce json
// @Param session_id path string true "会话ID"
// @Param q query string true "搜索关键词"
// @Param limit query int false "返回数量" default(10)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /chat/recalls/{session_id} [get]
// @Security BearerAuth
func (h *Handler) getRecalls(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		response.Err(c, response.CodeInvalidParams, "session_id is required")
		return
	}

	query := c.Query("q")
	if query == "" {
		response.Err(c, response.CodeInvalidParams, "query parameter 'q' is required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	pattern := "%" + query + "%"

	recalls, err := h.repo.SearchChatRecall(c.Request.Context(), repo.SearchChatRecallParams{
		SessionID: sessionID,
		Query:     pattern,
		Limit:     int32(limit),
	})
	if err != nil {
		h.logger.Errorf("[chat] get recalls failed: %v", err)
		response.ErrWithInternal(c, response.CodeDatabaseError, "failed to get recalls", err)
		return
	}

	response.Ok(c, gin.H{"recalls": recalls})
}
