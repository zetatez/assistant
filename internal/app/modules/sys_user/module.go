package sys_user

import (
	"assistant/internal/app/module"

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

func (m *SysUserModule) Name() string { return "sys_user" }

func (m *SysUserModule) Register(r *gin.Engine) {
	m.handler.Register(r.Group("/" + m.Name()))
}
