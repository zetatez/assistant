package task_def

import (
	"assistant/internal/app/repo"
	"assistant/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskDefHandler struct {
	svc *TaskDefService
}

func NewTaskDefHandler(svc *TaskDefService) *TaskDefHandler {
	return &TaskDefHandler{svc: svc}
}

func (h *TaskDefHandler) Register(r *gin.RouterGroup) {
	r.POST("/count", h.CountTaskDefs)
	r.POST("/create", h.CreateTaskDef)
	r.DELETE("/delete/:id", h.DeleteTaskDef)
	r.GET("/get/:id", h.GetTaskDefByID)
	r.GET("/list", h.ListTaskDefs)
	r.POST("/update", h.UpdateTaskDefByID)
}

func (h *TaskDefHandler) CountTaskDefs(c *gin.Context) {
	data, err := h.svc.CountTaskDefs(c)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskDefHandler) CreateTaskDef(c *gin.Context) {
	var req repo.CreateTaskDefParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.CreateTaskDef(c, req); err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

func (h *TaskDefHandler) DeleteTaskDef(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.DeleteTaskDefByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskDefHandler) GetTaskDefByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.GetTaskDefByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskDefHandler) ListTaskDefs(c *gin.Context) {
	var req repo.ListTaskDefsParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.ListTaskDefs(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskDefHandler) UpdateTaskDefByID(c *gin.Context) {
	var req repo.UpdateTaskDefByIDParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.UpdateTaskDefByID(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}
