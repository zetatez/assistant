package task_def_relation

import (
	"assistant/internal/app/module"

	"github.com/gin-gonic/gin"
)

type TaskDefRelationModule struct {
	handler *TaskDefRelationHandler
}

func NewTaskDefRelationModule() module.Module {
	return &TaskDefRelationModule{
		handler: NewTaskDefRelationHandler(NewTaskDefRelationService()),
	}
}

func (m *TaskDefRelationModule) Name() string { return "task_def_relation" }

func (m *TaskDefRelationModule) Register(r *gin.Engine) {
	m.handler.Register(r.Group("/" + m.Name()))
}
