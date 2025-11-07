package task_def

import (
	"assistant/internal/app/module"

	"github.com/gin-gonic/gin"
)

type TaskDefModule struct {
	handler *TaskDefHandler
}

func NewTaskDefModule() module.Module {
	return &TaskDefModule{
		handler: NewTaskDefHandler(NewTaskDefService()),
	}
}

func (m *TaskDefModule) Name() string { return "task_def" }

func (m *TaskDefModule) Register(r *gin.Engine) {
	m.handler.Register(r.Group("/" + m.Name()))
}
