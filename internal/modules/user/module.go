package user

import (
	"assistant/internal/db"
	"assistant/internal/model"
	"assistant/internal/module"

	"github.com/gin-gonic/gin"
)

type UserModule struct {
	handler *Handler
}

func NewUserModule() module.Module {
	return &UserModule{
		handler: NewHandler(NewService(NewRepo())),
	}
}

func (m *UserModule) Name() string { return "user" }

func (m *UserModule) Register(r *gin.Engine) {
	m.handler.Register(r.Group("/users"))
}

func (m *UserModule) Migrate() error {
	return db.DB.AutoMigrate(&model.User{})
}
