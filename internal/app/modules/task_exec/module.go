package task_exec

import (
	"assistant/internal/app/module"

	"github.com/gin-gonic/gin"
)

type TaskExecModule struct {
	handler *TaskExecHandler
}

func NewTaskExecModule() module.Module {
	return &TaskExecModule{
		handler: NewTaskExecHandler(NewTaskExecService()),
	}
}

func (m *TaskExecModule) Name() string { return "task_exec" }

func (m *TaskExecModule) Register(r *gin.Engine) {
	m.handler.Register(r.Group("/" + m.Name()))
}
