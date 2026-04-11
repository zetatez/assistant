package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	_ "assistant/docs"
	"assistant/internal/app/module"
	"assistant/internal/app/modules/chat"
	"assistant/internal/app/modules/health"
	"assistant/internal/app/modules/sys_server"
	"assistant/internal/app/modules/sys_user"
	"assistant/internal/app/modules/tars"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/channel"
	channel_feishu "assistant/pkg/channel/feishu"
	"assistant/pkg/middleware"
	"assistant/pkg/monitor/metrics"
	"assistant/pkg/tracing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const APIVersion = "v1"

func createChannel(cfg *psl.Config) (channel.Channel, error) {
	provider := cfg.Channel.Provider
	switch provider {
	case "feishu":
		feishuCfg := cfg.Channel.Feishu
		if feishuCfg.AppID == "" || feishuCfg.AppSecret == "" {
			return nil, fmt.Errorf("feishu app_id or app_secret is empty")
		}
		return channel_feishu.NewModule(feishuCfg.AppID, feishuCfg.AppSecret).AsChannel(), nil
	default:
		return nil, fmt.Errorf("unknown channel provider: %s", provider)
	}
}

func setupChannel(botChannel channel.Channel, cfg *psl.Config, logger *logrus.Logger) {
	logger.Info("setupChannel: entered")
	if !cfg.Tars.Enabled || botChannel == nil {
		logger.Info("setupChannel: early return")
		return
	}
	logger.Info("setupChannel: starting listener goroutine")
	go func() {
		logger.Info("tars: starting channel listener")
		if err := botChannel.StartListening(context.Background()); err != nil {
			logger.Errorf("tars: channel listener error: %v", err)
		}
		logger.Info("tars: channel listener goroutine ended")
	}()
	logger.Info("setupChannel: goroutine launched, returning")
}

func shutdown(botChannel channel.Channel) {
	middleware.StopLimiter()
	if botChannel != nil {
		botChannel.StopListening()
	}
}

func Run(ctx context.Context) error {
	logger := psl.GetLogger()
	cfg := psl.GetConfig()

	tracing.Init(cfg.Monitor.Tracing.Enabled, cfg.Monitor.Tracing.SampleRate)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(tracing.Middleware())
	r.Use(metrics.Middleware())

	r.POST("/auth/login", sys_user.NewAuthHandler().Login)

	if cfg.Monitor.Metrics.Enabled {
		r.GET(cfg.Monitor.Metrics.Path, func(c *gin.Context) {
			c.Header("Content-Type", "text/plain")
			c.String(200, metrics.FormatPrometheus())
		})
	}

	botChannel, err := createChannel(cfg)
	if err != nil {
		logger.WithFields(map[string]interface{}{"error": err.Error()}).Warnf("channel creation failed, tars disabled")
		botChannel = nil
	}

	setupChannel(botChannel, cfg, logger)

	modules := []module.Module{
		health.NewHealthModule(),
		sys_server.NewSysServerModule(),
		sys_user.NewSysUserModule(),
		chat.NewChatModule(),
	}

	if cfg.Tars.Enabled && botChannel != nil {
		tarsModule := tars.NewModule(ctx, botChannel)
		if tarsModule.Name() != "" {
			modules = append(modules, tarsModule)
		}
	}

	apiV1 := r.Group("/api/" + APIVersion)
	apiV1.Use(middleware.RateLimit(100, 60))
	{
		for _, m := range modules {
			logger.WithFields(map[string]interface{}{"module": m.Name(), "prefix": "/api/" + APIVersion + "/" + m.Name(), "version": APIVersion}).Info("registering module")
			group := apiV1.Group("/" + m.Name())
			moduleMiddleware := m.Middleware()
			if len(moduleMiddleware) > 0 {
				group.Use(moduleMiddleware...)
			}
			m.Register(group)
		}
	}

	logger.WithFields(map[string]interface{}{"port": psl.GetConfig().App.Port, "swagger": "/swagger/index.html"}).Info("swagger documentation available")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	addr := fmt.Sprintf(":%d", psl.GetConfig().App.Port)
	logger.WithFields(map[string]interface{}{"address": addr, "version": APIVersion}).Info("server running")

	srv := &http.Server{Addr: addr, Handler: r}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.WithFields(map[string]interface{}{"reason": ctx.Err().Error()}).Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdown(botChannel)
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		logger.WithFields(map[string]interface{}{"reason": err.Error()}).Info("server error received")
		shutdown(botChannel)
		return err
	}
}
