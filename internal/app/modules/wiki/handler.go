package wiki

import (
	"assistant/internal/app/repo"
	"assistant/pkg/response"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo   *repo.WikiRepo
	logger Logger
}

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
}

func NewHandler(repo *repo.WikiRepo, logger Logger) *Handler {
	return &Handler{
		repo:   repo,
		logger: logger,
	}
}

func (h *Handler) Register(r *gin.RouterGroup) {
	r.POST("/entries", h.createEntry)
	r.GET("/entries", h.listEntries)
	r.GET("/entries/:id", h.getEntry)
	r.DELETE("/entries/:id", h.deleteEntry)
	r.GET("/search", h.search)
	r.POST("/import", h.importFile)
}

// createEntry godoc
// @Summary 创建知识库条目
// @Description 创建新的知识库条目
// @Tags 知识库
// @Accept json
// @Produce json
// @Param data body createEntryRequest true "条目信息"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /wiki/entries [post]
func (h *Handler) createEntry(c *gin.Context) {
	var req struct {
		Title    string `json:"title" binding:"required"`
		Content  string `json:"content" binding:"required"`
		Keywords string `json:"keywords"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid request parameters")
		return
	}

	entry, err := h.repo.CreateEntry(c.Request.Context(), req.Title, req.Content, req.Keywords, "admin")
	if err != nil {
		h.logger.Errorf("[wiki] create entry failed: %v", err)
		response.ErrWithInternal(c, response.CodeDatabaseError, "failed to create entry", err)
		return
	}

	response.Ok(c, entry)
}

// listEntries godoc
// @Summary 获取知识库条目列表
// @Description 分页获取知识库条目列表
// @Tags 知识库
// @Produce json
// @Param offset query int false "偏移量" default(0)
// @Param limit query int false "返回数量" default(20)
// @Success 200 {object} response.Response "成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /wiki/entries [get]
func (h *Handler) listEntries(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	entries, err := h.repo.List(c.Request.Context(), offset, limit)
	if err != nil {
		h.logger.Errorf("[wiki] list entries failed: %v", err)
		response.ErrWithInternal(c, response.CodeDatabaseError, "failed to list entries", err)
		return
	}

	response.Ok(c, entries)
}

// getEntry godoc
// @Summary 获取知识库条目详情
// @Description 根据ID获取知识库条目详情
// @Tags 知识库
// @Produce json
// @Param id path int true "条目ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "条目不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /wiki/entries/{id} [get]
func (h *Handler) getEntry(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid entry id")
		return
	}

	entry, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Errorf("[wiki] get entry failed: %v", err)
		response.ErrWithInternal(c, response.CodeDatabaseError, "failed to get entry", err)
		return
	}

	if entry == nil {
		response.Err(c, response.CodeNotFound, "entry not found")
		return
	}

	response.Ok(c, entry)
}

// deleteEntry godoc
// @Summary 删除知识库条目
// @Description 根据ID删除知识库条目
// @Tags 知识库
// @Produce json
// @Param id path int true "条目ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /wiki/entries/{id} [delete]
func (h *Handler) deleteEntry(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, response.CodeInvalidParams, "invalid entry id")
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		h.logger.Errorf("[wiki] delete entry failed: %v", err)
		response.ErrWithInternal(c, response.CodeDatabaseError, "failed to delete entry", err)
		return
	}

	response.Ok(c, nil)
}

// search godoc
// @Summary 搜索知识库条目
// @Description 根据关键词搜索知识库条目
// @Tags 知识库
// @Produce json
// @Param q query string true "搜索关键词"
// @Param limit query int false "返回数量" default(5)
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /wiki/search [get]
func (h *Handler) search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Err(c, response.CodeInvalidParams, "missing query parameter 'q'")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	entries, err := h.repo.Search(c.Request.Context(), query, limit)
	if err != nil {
		h.logger.Errorf("[wiki] search failed: %v", err)
		response.ErrWithInternal(c, response.CodeDatabaseError, "failed to search entries", err)
		return
	}

	response.Ok(c, gin.H{"entries": entries})
}

// importFile godoc
// @Summary 导入知识库条目
// @Description 从Markdown文件导入知识库条目
// @Tags 知识库
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Markdown文件"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /wiki/import [post]
func (h *Handler) importFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Err(c, response.CodeInvalidParams, "no file uploaded")
		return
	}

	if filepath.Ext(file.Filename) != ".md" {
		response.Err(c, response.CodeInvalidParams, "only .md files supported")
		return
	}

	f, err := file.Open()
	if err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to open file", err)
		return
	}
	defer f.Close()

	content := make([]byte, file.Size)
	if _, err := f.Read(content); err != nil {
		response.ErrWithInternal(c, response.CodeServerError, "failed to read file", err)
		return
	}

	title := file.Filename[:len(file.Filename)-3]
	entry, err := h.repo.CreateEntry(c.Request.Context(), title, string(content), "", "admin")
	if err != nil {
		h.logger.Errorf("[wiki] import entry failed: %v", err)
		response.ErrWithInternal(c, response.CodeDatabaseError, "failed to import entry", err)
		return
	}

	response.Ok(c, entry)
}

type createEntryRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Keywords string `json:"keywords"`
}
