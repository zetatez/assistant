package health

import (
	"net/http"

	"assistant/pkg/response"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	svc *HealthService
}

func NewHealthHandler(svc *HealthService) *HealthHandler {
	return &HealthHandler{svc: svc}
}

func (h *HealthHandler) Register(r *gin.RouterGroup) {
	r.GET("", h.Health)
}

func (h *HealthHandler) Health(c *gin.Context) {
	data, err := h.svc.Health()
	if err != nil {
		response.Err(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(c, data)
}
