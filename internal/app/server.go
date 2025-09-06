package app

import (
	_ "assistant/docs"
	"assistant/internal/app/module"
	"assistant/internal/app/modules/health"
	"assistant/internal/app/modules/task_orchestration"
	"assistant/internal/app/modules/user"
	"assistant/internal/cfg"
	"assistant/internal/log"
	"fmt"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	modules := []module.Module{
		health.NewHealthModule(),
		user.NewUserModule(),
		task_orchestration.NewTaskOrchestrationModule(),
	}

	for _, m := range modules {
		log.Logger.Printf("✅ - register module %s", m.Name())
		m.Register(r)
	}

	log.Logger.Infof("✅ swag on: http://127.0.0.1:%d/swagger/index.html", cfg.C.App.Port)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Logger.Infof("✅ Server running at :%d", cfg.C.App.Port)
	r.Run(fmt.Sprintf(":%d", cfg.C.App.Port))
}
