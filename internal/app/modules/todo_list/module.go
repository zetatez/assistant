package todo_list

import (
	"assistant/internal/app/module"

	"github.com/gin-gonic/gin"
)

type TodoListModule struct {
	handler *TodoListHandler
}

func NewTodoListModule() module.Module {
	return &TodoListModule{
		handler: NewTodoListHandler(NewTodoListService()),
	}
}

func (m *TodoListModule) Name() string { return "todo_list" }

func (m *TodoListModule) Register(r *gin.Engine) {
	m.handler.Register(r.Group("/" + m.Name()))
}
