package bootstrap

import (
	"assistant/internal/app"
	"assistant/internal/bootstrap/psl"
	"context"
	"fmt"
	"time"
)

func Run(ctx context.Context) error {
	if err := psl.InitConfig(); err != nil {
		return fmt.Errorf("init config failed: %w", err)
	}

	if err := psl.InitLog(); err != nil {
		return fmt.Errorf("init log failed: %w", err)
	}

	logger := psl.GetLogger()
	logger.Println("init log success")

	if err := psl.InitDB(ctx); err != nil {
		return fmt.Errorf("init db failed: %w", err)
	}
	logger.Info("init db success")

	defer func() {
		if db := psl.GetDB(); db != nil {
			db.Close()
		}
	}()

	if err := psl.InitDisLocker(ctx); err != nil {
		return fmt.Errorf("init distributed locker failed: %w", err)
	}
	psl.GetDisLocker().StartExpiredLockCleaner(ctx, 15*time.Minute)
	logger.Info("init distributed locker success")

	if err := psl.MigrateDB(ctx); err != nil {
		return fmt.Errorf("migrate db failed: %w", err)
	}
	logger.Info("migrate db success")

	return app.Run(ctx)
}
