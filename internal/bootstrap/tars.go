package bootstrap

import (
	"context"

	"assistant/internal/bootstrap/psl"
	"assistant/internal/tars"
	"assistant/pkg/channel/feishu"
)

func InitTars(ctx context.Context) *tars.Service {
	cfg := psl.GetConfig()
	if !cfg.Tars.Enabled {
		return nil
	}

	logger := psl.GetLogger()

	if cfg.Settings.Feishu.AppID == "" || cfg.Settings.Feishu.AppSecret == "" {
		logger.Error("tars: feishu app_id or app_secret is empty, skipping tars module")
		return nil
	}

	ch := feishu.NewService(cfg.Settings.Feishu.AppID, cfg.Settings.Feishu.AppSecret, feishu.WithLogger(logger))

	tarsCfg := &tars.TarsConfig{
		Enabled:          cfg.Tars.Enabled,
		DataDir:          cfg.Tars.DataDir,
		Temperature:      cfg.Tars.Temperature,
		MaxHistoryMB:     cfg.Tars.Memory.MaxHistoryMB,
		SummarizeEvery:   cfg.Tars.Memory.SummarizeEvery,
		MaxShortTerm:     cfg.Tars.Memory.MaxShortTerm,
		WikiEnabled:      cfg.Tars.Wiki.Enabled,
		WikiDir:          cfg.Tars.Wiki.Dir,
		WebSearchEnabled: cfg.Tars.WebSearch.Enabled,
	}

	svc := tars.NewService(ch, psl.GetLLMClient(), cfg.LLMProxy.ProxiedModel, tarsCfg, logger)

	svc.Start(ctx)

	psl.RegisterCleanup(func(_ context.Context) {
		svc.Stop()
	})

	logger.Info("tars: service initialized and started")
	return svc
}
