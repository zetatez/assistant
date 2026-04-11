package chat

import (
	"assistant/internal/app/module"
	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type ChatModule struct {
	name    string
	handler *Handler
}

func NewChatModule() module.Module {
	db := psl.GetDB()
	if db == nil {
		return &ChatModule{name: ""}
	}

	queries := repo.New(db)

	return &ChatModule{
		name:    "chat",
		handler: NewHandler(queries, psl.GetLogger()),
	}
}

func (m *ChatModule) Name() string { return m.name }

func (m *ChatModule) Register(r *gin.RouterGroup) {
	if m.handler == nil {
		return
	}
	m.handler.Register(r)
}

func (m *ChatModule) Middleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.AuthRequired(psl.GetConfig().Auth.JWT.Secret),
	}
}
