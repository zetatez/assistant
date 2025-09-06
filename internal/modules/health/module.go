package health

import (
	"assistant/internal/module"

	"github.com/gin-gonic/gin"
)

type HealthModule struct {
	handler *Handler
}

func NewHealthModule() module.Module {
	return &HealthModule{
		handler: NewHandler(NewService(NewRepo())),
	}
}

func (m *HealthModule) Name() string { return "health" }

func (m *HealthModule) Register(r *gin.Engine) {
	m.handler.Register(r.Group("/health"))
}

func (m *HealthModule) Migrate() error {
	return nil
}
