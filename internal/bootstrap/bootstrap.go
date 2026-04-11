package bootstrap

import (
	"context"
	"fmt"
	"time"

	"assistant/internal/app"
	"assistant/internal/bootstrap/psl"

	_ "assistant/pkg/llm/providers/deepseek"
	_ "assistant/pkg/llm/providers/doubao"
	_ "assistant/pkg/llm/providers/gemini"
	_ "assistant/pkg/llm/providers/glm"
	_ "assistant/pkg/llm/providers/minimax"
	_ "assistant/pkg/llm/providers/openai"
)

func Run(ctx context.Context) error {
	if err := psl.InitConfig(); err != nil {
		return fmt.Errorf("init config failed: %w", err)
	}

	if err := psl.InitLog(); err != nil {
		return fmt.Errorf("init log failed: %w", err)
	}

	logger := psl.GetLogger()
	logger.Info("init log success")

	if err := psl.InitDB(ctx); err != nil {
		return fmt.Errorf("init db failed: %w", err)
	}
	logger.Info("init db success")

	if err := psl.MigrateDB(ctx); err != nil {
		return fmt.Errorf("migrate db failed: %w", err)
	}
	logger.Info("migrate db success")

	svrIP, err := psl.EnsureLocalSysServerRegistered(ctx)
	if err != nil {
		return fmt.Errorf("register local sys_server failed: %w", err)
	}
	logger.Infof("local sys_server registered: %s", svrIP)
	psl.StartSysServerMonitor(ctx, svrIP, 15*time.Second)

	defer func() {
		psl.ShutdownAll()
		if db := psl.GetDB(); db != nil {
			db.Close()
		}
	}()

	return app.Run(ctx)
}
