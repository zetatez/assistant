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
	r.POST("/create", h.CreateTodoList)
	r.GET("/get/:id", h.GetTodoList)
	r.GET("/done", h.DoneTodoList)
	r.GET("/delete", h.DeleteTodoList)
}

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

func (h *TodoListHandler) GetTodoList(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.GetTodoList(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TodoListHandler) DoneTodoList(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	err = h.svc.DoneTodoList(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

func (h *TodoListHandler) DeleteTodoList(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	err = h.svc.DeleteTodoList(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}
