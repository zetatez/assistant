package app

import (
	_ "assistant/docs"
	"assistant/internal/app/module"
	"assistant/internal/app/modules/health"
	"assistant/internal/app/modules/sys_user"
	"assistant/internal/app/modules/task_orchestration"
	"assistant/internal/bootstrap/psl"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run(ctx context.Context) error {
	logger := psl.GetLogger()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	modules := []module.Module{
		health.NewHealthModule(),
		sys_user.NewSysUserModule(),
		task_orchestration.NewTaskOrchestrationModule(),
	}

	for _, m := range modules {
		logger.Infof("register module %s", m.Name())
		m.Register(r)
	}

	logger.Infof("swag on: http://127.0.0.1:%d/swagger/index.html", psl.GetConfig().App.Port)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	addr := fmt.Sprintf(":%d", psl.GetConfig().App.Port)
	logger.Infof("server running at %s", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Infof("shutdown signal received: %v", ctx.Err())
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
