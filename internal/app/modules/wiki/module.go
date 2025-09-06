package wiki

import (
	"assistant/internal/app/module"
	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type WikiModule struct {
	name    string
	handler *Handler
}

func NewWikiModule() module.Module {
	db := psl.GetDB()
	if db == nil {
		return &WikiModule{name: ""}
	}

	wikiRepo := repo.NewWikiRepo(repo.New(db))

	return &WikiModule{
		name:    "wiki",
		handler: NewHandler(wikiRepo, psl.GetLogger()),
	}
}

func (m *WikiModule) Name() string { return m.name }

func (m *WikiModule) Register(r *gin.RouterGroup) {
	if m.handler == nil {
		return
	}
	m.handler.Register(r)
}

func (m *WikiModule) Middleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.AuthRequired(psl.GetConfig().App.JWT.Secret),
	}
}
