package todo_list

import (
	"assistant/internal/app/repo"
	"assistant/pkg/response"
	"database/sql"
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
	r.GET("/count", h.CountTodoList)
	r.POST("/create", h.CreateTodoList)
	r.DELETE("/delete/:id", h.DeleteTodoList)
	r.GET("/get/:id", h.GetTodoListByID)
	r.GET("/list", h.ListTodoLists)
	r.POST("/search_by_content", h.SearchTodoListsByContent)
	r.POST("/search_by_deadline_lt", h.SearchTodoListsByDeadlineLT)
	r.POST("/search_by_title", h.SearchTodoListsByTitle)
	r.POST("/search_by_title_and_content", h.SearchTodoListsByTitleAndContent)
	r.PUT("/update/:id", h.UpdateTodoListByID)
	r.PATCH("/progress/:id", h.UpdateTodoListProgressByID)
	r.POST("/complete/:id", h.CompleteTodoListByID)
	r.PATCH("/priority/:id", h.UpdateTodoListPriorityByID)
}

// CountTodoList godoc
// @Summary 获取待办事项总数
// @Description 统计系统中所有 Todo 的数量
// @Tags 待办事项
// @Produce json
// @Success 200 {object} response.Response "成功"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/count [get]
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
// @Param id path int true "待办事项ID"
// @Param data body repo.UpdateTodoListByIDParams true "更新参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/update/{id} [put]
func (h *TodoListHandler) UpdateTodoListByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	var req repo.UpdateTodoListByIDParams
	req.ID = sql.NullInt64{Int64: id, Valid: true}
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

// UpdateTodoListProgressByID godoc
// @Summary 更新待办事项进度
// @Description 根据 ID 更新 Todo 进度，进度100时自动设为COMPLETED
// @Tags 待办事项
// @Accept json
// @Produce json
// @Param id path int true "待办事项ID"
// @Param progress query int true "进度 0-100"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/progress/{id} [patch]
func (h *TodoListHandler) UpdateTodoListProgressByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	progress, err := strconv.Atoi(c.Query("progress"))
	if err != nil || progress < 0 || progress > 100 {
		response.Err(c, http.StatusBadRequest, "progress must be between 0 and 100")
		return
	}
	err = h.svc.UpdateTodoListProgressByID(c, id, int64(progress))
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// CompleteTodoListByID godoc
// @Summary 完成待办事项
// @Description 根据 ID 将 Todo 进度设为100，状态设为COMPLETED
// @Tags 待办事项
// @Produce json
// @Param id path int true "待办事项ID"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/complete/{id} [post]
func (h *TodoListHandler) CompleteTodoListByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	err = h.svc.CompleteTodoListByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

// UpdateTodoListPriorityByID godoc
// @Summary 更新待办事项优先级
// @Description 根据 ID 更新 Todo 优先级 (1-10, 越高越紧急)
// @Tags 待办事项
// @Accept json
// @Produce json
// @Param id path int true "待办事项ID"
// @Param priority query int true "优先级 1-10"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /todo_list/priority/{id} [patch]
func (h *TodoListHandler) UpdateTodoListPriorityByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	priority, err := strconv.Atoi(c.Query("priority"))
	if err != nil || priority < 1 || priority > 10 {
		response.Err(c, http.StatusBadRequest, "priority must be between 1 and 10")
		return
	}
	err = h.svc.UpdateTodoListPriorityByID(c, id, int64(priority))
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}
