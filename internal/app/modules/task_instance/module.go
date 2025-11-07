package task_instance

import (
	"assistant/internal/app/module"

	"github.com/gin-gonic/gin"
)

type TaskInstanceModule struct {
	handler *TaskInstanceHandler
}

func NewTaskInstanceModule() module.Module {
	return &TaskInstanceModule{
		handler: NewTaskInstanceHandler(NewTaskInstanceService()),
	}
}

func (m *TaskInstanceModule) Name() string { return "task_instance" }

func (m *TaskInstanceModule) Register(r *gin.Engine) {
	m.handler.Register(r.Group("/" + m.Name()))
}
