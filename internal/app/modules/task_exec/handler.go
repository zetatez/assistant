package task_exec

import (
	"assistant/internal/app/repo"
	"assistant/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskExecHandler struct {
	svc *TaskExecService
}

func NewTaskExecHandler(svc *TaskExecService) *TaskExecHandler {
	return &TaskExecHandler{svc: svc}
}

func (h *TaskExecHandler) Register(r *gin.RouterGroup) {
	r.POST("/count", h.CountTaskExecs)
	r.POST("/create", h.CreateTaskExec)
	r.DELETE("/delete/:id", h.DeleteTaskExec)
	r.GET("/get/:id", h.GetTaskExecByID)
	r.GET("/list", h.ListTaskExecs)
	r.POST("/update", h.UpdateTaskExecByID)
}

func (h *TaskExecHandler) CountTaskExecs(c *gin.Context) {
	data, err := h.svc.CountTaskExecs(c)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskExecHandler) CreateTaskExec(c *gin.Context) {
	var req repo.CreateTaskExecParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.CreateTaskExec(c, req); err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

func (h *TaskExecHandler) DeleteTaskExec(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.DeleteTaskExecByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskExecHandler) GetTaskExecByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.GetTaskExecByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskExecHandler) ListTaskExecs(c *gin.Context) {
	var req repo.ListTaskExecsParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.ListTaskExecs(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskExecHandler) UpdateTaskExecByID(c *gin.Context) {
	var req repo.UpdateTaskExecByIDParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.UpdateTaskExecByID(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}
