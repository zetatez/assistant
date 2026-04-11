package sys_user

import (
	"assistant/internal/app/module"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type SysUserModule struct {
	handler *SysUserHandler
}

func NewSysUserModule() module.Module {
	return &SysUserModule{
		handler: NewSysUserHandler(NewSysUserService()),
	}
}

func NewSysUserModuleWithoutAuth() *SysUserModule {
	return &SysUserModule{
		handler: NewSysUserHandler(NewSysUserService()),
	}
}

func (m *SysUserModule) Name() string { return "sys_user" }

func (m *SysUserModule) Register(r *gin.RouterGroup) {
	m.handler.RegisterWithoutAuth(r)
}

func (m *SysUserModule) Middleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.AuthRequired(psl.GetConfig().Auth.JWT.Secret),
	}
}

func (m *SysUserModule) GetAuthHandler() *SysUserHandler {
	return m.handler
}

func NewAuthHandler() *SysUserHandler {
	return NewSysUserHandler(NewSysUserService())
}
