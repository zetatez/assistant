package app

import (
	"assistant/internal/app/module"
	"assistant/internal/app/modules/health"
	"assistant/internal/app/modules/user"
	"log"

	"github.com/gin-gonic/gin"
)

func Run() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// modules
	modules := []module.Module{
		health.NewHealthModule(),
		user.NewUserModule(),
	}

	for _, m := range modules {
		log.Printf("registering module: %s", m.Name())
		m.Register(r)
	}
	// r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("Server running at :8080")
	r.Run(":8080")
}
