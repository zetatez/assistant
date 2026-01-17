package app

import (
	_ "assistant/docs"
	"assistant/internal/app/module"
	"assistant/internal/app/modules/health"
	"assistant/internal/app/modules/sys_user"
	"assistant/internal/app/modules/task_orchestration"
	"assistant/internal/bootstrap/psl"
	"fmt"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run() {
	logger := psl.GetLogger()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	modules := []module.Module{
		health.NewHealthModule(),
		sys_user.NewSysUserModule(),
		task_orchestration.NewTaskOrchestrationModule(),
	}

	for _, m := range modules {
		logger.Printf("register module %s", m.Name())
		m.Register(r)
	}

	logger.Infof("swag on: http://127.0.0.1:%d/swagger/index.html", psl.GetConfig().App.Port)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	logger.Infof("server running at :%d", psl.GetConfig().App.Port)
	r.Run(fmt.Sprintf(":%d", psl.GetConfig().App.Port))
}
