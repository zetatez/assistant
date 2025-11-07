package task_instance

import (
	"assistant/internal/app/repo"
	"assistant/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskInstanceHandler struct {
	svc *TaskInstanceService
}

func NewTaskInstanceHandler(svc *TaskInstanceService) *TaskInstanceHandler {
	return &TaskInstanceHandler{svc: svc}
}

func (h *TaskInstanceHandler) Register(r *gin.RouterGroup) {
	r.POST("/count", h.CountTaskInstances)
	r.POST("/create", h.CreateTaskInstance)
	r.DELETE("/delete/:id", h.DeleteTaskInstance)
	r.GET("/get/:id", h.GetTaskInstanceByID)
	r.GET("/list", h.ListTaskInstances)
	r.POST("/update", h.UpdateTaskInstanceByID)
}

func (h *TaskInstanceHandler) CountTaskInstances(c *gin.Context) {
	data, err := h.svc.CountTaskInstances(c)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskInstanceHandler) CreateTaskInstance(c *gin.Context) {
	var req repo.CreateTaskInstanceParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.CreateTaskInstance(c, req); err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

func (h *TaskInstanceHandler) DeleteTaskInstance(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.DeleteTaskInstanceByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskInstanceHandler) GetTaskInstanceByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.GetTaskInstanceByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskInstanceHandler) ListTaskInstances(c *gin.Context) {
	var req repo.ListTaskInstancesParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.ListTaskInstances(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskInstanceHandler) UpdateTaskInstanceByID(c *gin.Context) {
	var req repo.UpdateTaskInstanceByIDParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.UpdateTaskInstanceByID(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}
