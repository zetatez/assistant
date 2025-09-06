package module

import "github.com/gin-gonic/gin"

type Module interface {
	Name() string
	Register(r *gin.Engine)
	Migrate() error
}
