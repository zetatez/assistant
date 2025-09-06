package user

import (
	"assistant/internal/app/module"

	"github.com/gin-gonic/gin"
)

type UserModule struct {
	handler *UserHandler
}

func NewUserModule() module.Module {
	return &UserModule{
		handler: NewUserHandler(NewUserService()),
	}
}

func (m *UserModule) Name() string { return "user" }

func (m *UserModule) Register(r *gin.Engine) {
	m.handler.Register(r.Group("/" + m.Name()))
}
