package sys_distributed_locker

import (
	"assistant/internal/app/module"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type SysDistributedLockModule struct {
	handler *SysDistributedLockHandler
}

func NewSysDistributedLockModule() module.Module {
	return &SysDistributedLockModule{
		handler: NewSysDistributedLockHandler(NewSysDistributedLockService()),
	}
}

func (m *SysDistributedLockModule) Name() string { return "sys_distributed_locker" }

func (m *SysDistributedLockModule) Register(r *gin.RouterGroup) {
	m.handler.Register(r)
}

func (m *SysDistributedLockModule) Middleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.AuthRequired(psl.GetConfig().App.JWT.Secret),
	}
}
