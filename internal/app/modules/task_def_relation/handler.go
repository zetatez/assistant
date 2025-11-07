package task_def_relation

import (
	"assistant/internal/app/repo"
	"assistant/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskDefRelationHandler struct {
	svc *TaskDefRelationService
}

func NewTaskDefRelationHandler(svc *TaskDefRelationService) *TaskDefRelationHandler {
	return &TaskDefRelationHandler{svc: svc}
}

func (h *TaskDefRelationHandler) Register(r *gin.RouterGroup) {
	r.POST("/count", h.CountTaskDefRelations)
	r.POST("/create", h.CreateTaskDefRelation)
	r.DELETE("/delete/:id", h.DeleteTaskDefRelation)
	r.GET("/get/:id", h.GetTaskDefRelationByID)
	r.GET("/list", h.ListTaskDefRelations)
	r.POST("/update", h.UpdateTaskDefRelationByID)
}

func (h *TaskDefRelationHandler) CountTaskDefRelations(c *gin.Context) {
	data, err := h.svc.CountTaskDefRelations(c)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskDefRelationHandler) CreateTaskDefRelation(c *gin.Context) {
	var req repo.CreateTaskDefRelationParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.CreateTaskDefRelation(c, req); err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, nil)
}

func (h *TaskDefRelationHandler) DeleteTaskDefRelation(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.DeleteTaskDefRelationByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskDefRelationHandler) GetTaskDefRelationByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.GetTaskDefRelationByID(c, id)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskDefRelationHandler) ListTaskDefRelations(c *gin.Context) {
	var req repo.ListTaskDefRelationsParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.ListTaskDefRelations(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}

func (h *TaskDefRelationHandler) UpdateTaskDefRelationByID(c *gin.Context) {
	var req repo.UpdateTaskDefRelationByIDParams
	if err := c.BindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.svc.UpdateTaskDefRelationByID(c, req)
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}
