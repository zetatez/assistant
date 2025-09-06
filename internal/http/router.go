package httpserver

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/zetatez/assistant/internal/config"
	"github.com/zetatez/assistant/internal/repository"
	"github.com/zetatez/assistant/internal/service"
	"github.com/zetatez/assistant/pkg/response"
)

func NewRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	if cfg.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	r.Use(cors.Default())

	// health check
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, response.OK(gin.H{"status": "healthy"})) })

	// DI
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo)

	api := r.Group("/api")
	{
		api.POST("/users", func(c *gin.Context) { createUser(c, userSvc) })
		api.GET("/users/:id", func(c *gin.Context) { getUser(c, userSvc) })
		api.GET("/users", func(c *gin.Context) { listUsers(c, userSvc) })
		api.PUT("/users/:id", func(c *gin.Context) { updateUser(c, userSvc) })
		api.DELETE("/users/:id", func(c *gin.Context) { deleteUser(c, userSvc) })
	}
	return r
}
