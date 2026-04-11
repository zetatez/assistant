package sys_server

import (
	"assistant/internal/app/module"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type SysServerModule struct {
	handler *SysServerHandler
}

func NewSysServerModule() module.Module {
	return &SysServerModule{
		handler: NewSysServerHandler(NewSysServerService()),
	}
}

func (m *SysServerModule) Name() string { return "sys_server" }

func (m *SysServerModule) Register(r *gin.RouterGroup) {
	m.handler.Register(r)
}

func (m *SysServerModule) Middleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.AuthRequired(psl.GetConfig().Auth.JWT.Secret),
	}
}
