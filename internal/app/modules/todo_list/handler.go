package todo_list

import (
	"assistant/internal/app/repo"
	"assistant/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TodoListHandler struct {
	svc *TodoListService
}

func NewTodoListHandler(svc *TodoListService) *TodoListHandler {
	return &TodoListHandler{svc: svc}
}

func (h *TodoListHandler) Register(r *gin.RouterGroup) {
	r.POST("/count", h.CountTodoList)
	r.POST("/create", h.CreateTodoList)
	r.DELETE("/delete/:id", h.DeleteTodoList)
	r.GET("/get/:id", h.GetTodoListByID)
	r.GET("/list", h.ListTodoLists)
	r.POST("/mark_done/:id", h.MarkTodoListAsDoneByID)
	r.POST("/search_by_content", h.SearchTodoListsByContent)
	r.POST("/search_by_deadline_lt", h.SearchTodoListsByDeadlineLT)
	r.POST("/search_by_title", h.SearchTodoListsByTitle)
	r.POST("/search_by_title_and_content", h.SearchTodoListsByTitleAndContent)
	r.POST("/update", h.UpdateTodoListByID)
}

// CountTodoList godoc
// @Summary 获取待办事项总数
// @Description 统计系统中所有 Todo 的数量
// @Tags 待办事项
// @Produce json
// @Success 200 {object} response.Response "成功"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/count [post]
func (h *TodoListHandler) CountTodoList(c *gin.Context) {
	data, err := h.svc.CountTodoList(c)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

// CreateTodoList godoc
// @Summary 创建待办事项
// @Description 创建一个新的 Todo 项
// @Tags 待办事项
// @Accept json
// @Produce json
// @Param data body repo.CreateTodoListParams true "创建参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/create [post]
func (h *TodoListHandler) CreateTodoList(c *gin.Context) {
	var req repo.CreateTodoListParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.CreateTodoList(c, req); err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// DeleteTodoList godoc
// @Summary 删除待办事项
// @Description 根据 ID 删除 Todo 项
// @Tags 待办事项
// @Produce json
// @Param id path int true "待办事项ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/delete/{id} [delete]
func (h *TodoListHandler) DeleteTodoList(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	err = h.svc.DeleteTodoListByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// GetTodoListByID godoc
// @Summary 获取待办事项详情
// @Description 根据 ID 获取单个 Todo 详情
// @Tags 待办事项
// @Produce json
// @Param id path int true "待办事项ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/get/{id} [get]
func (h *TodoListHandler) GetTodoListByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.GetTodoListByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

// ListTodoLists godoc
// @Summary 获取待办事项列表
// @Description 支持分页、排序的 Todo 列表查询
// @Tags 待办事项
// @Accept json
// @Produce json
// @Param data body repo.ListTodoListsParams true "查询参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/list [get]
func (h *TodoListHandler) ListTodoLists(c *gin.Context) {
	var req repo.ListTodoListsParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.ListTodoLists(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

// MarkTodoListAsDoneByID godoc
// @Summary 标记待办事项完成
// @Description 根据 ID 将 Todo 状态标记为完成
// @Tags 待办事项
// @Produce json
// @Param id path int true "待办事项ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/mark_done/{id} [post]
func (h *TodoListHandler) MarkTodoListAsDoneByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	err = h.svc.MarkTodoListAsDoneByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// SearchTodoListsByContent godoc
// @Summary 按内容搜索待办事项
// @Description 根据内容关键字模糊查询 Todo
// @Tags 待办事项
// @Accept json
// @Produce json
// @Param data body repo.SearchTodoListsByContentParams true "搜索参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/search_by_content [post]
func (h *TodoListHandler) SearchTodoListsByContent(c *gin.Context) {
	var req repo.SearchTodoListsByContentParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.SearchTodoListsByContent(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

// SearchTodoListsByDeadlineLT godoc
// @Summary 按截止时间查询待办事项
// @Description 查询截止时间早于指定时间的 Todo 项
// @Tags 待办事项
// @Accept json
// @Produce json
// @Param data body repo.SearchTodoListsByDeadlineLTParams true "搜索参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/search_by_deadline_lt [post]
func (h *TodoListHandler) SearchTodoListsByDeadlineLT(c *gin.Context) {
	var req repo.SearchTodoListsByDeadlineLTParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.SearchTodoListsByDeadlineLT(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

// SearchTodoListsByTitle godoc
// @Summary 按标题搜索待办事项
// @Description 根据标题关键字模糊查询 Todo
// @Tags 待办事项
// @Accept json
// @Produce json
// @Param data body repo.SearchTodoListsByTitleParams true "搜索参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/search_by_title [post]
func (h *TodoListHandler) SearchTodoListsByTitle(c *gin.Context) {
	var req repo.SearchTodoListsByTitleParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.SearchTodoListsByTitle(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

// SearchTodoListsByTitleAndContent godoc
// @Summary 按标题和内容搜索待办事项
// @Description 同时根据标题和内容进行模糊搜索
// @Tags 待办事项
// @Accept json
// @Produce json
// @Param data body repo.SearchTodoListsByTitleAndContentParams true "搜索参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/search_by_title_and_content [post]
func (h *TodoListHandler) SearchTodoListsByTitleAndContent(c *gin.Context) {
	var req repo.SearchTodoListsByTitleAndContentParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.SearchTodoListsByTitleAndContent(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

// UpdateTodoListByID godoc
// @Summary 更新待办事项
// @Description 根据 ID 更新 Todo 信息
// @Tags 待办事项
// @Accept json
// @Produce json
// @Param data body repo.UpdateTodoListByIDParams true "更新参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/update [post]
func (h *TodoListHandler) UpdateTodoListByID(c *gin.Context) {
	var req repo.UpdateTodoListByIDParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.UpdateTodoListByID(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}
