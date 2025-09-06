package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	_ "assistant/docs"
	"assistant/internal/app/module"
	"assistant/internal/app/modules/health"
	"assistant/internal/app/modules/sys_distributed_locker"
	"assistant/internal/app/modules/sys_server"
	"assistant/internal/app/modules/sys_user"
	"assistant/internal/app/modules/tars"
	"assistant/internal/app/modules/wiki"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/channel"
	channel_feishu "assistant/pkg/channel/feishu"
	"assistant/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const APIVersion = "v1"

func createChannel(cfg *psl.Config) (channel.Channel, error) {
	provider := cfg.Tars.ChannelProvider
	switch provider {
	case "feishu":
		feishuCfg := cfg.Channels.Feishu
		if feishuCfg.AppID == "" || feishuCfg.AppSecret == "" {
			return nil, fmt.Errorf("feishu app_id or app_secret is empty")
		}
		return channel_feishu.NewModule(feishuCfg.AppID, feishuCfg.AppSecret).AsChannel(), nil
	default:
		return nil, fmt.Errorf("unknown channel provider: %s", provider)
	}
}

func setupChannel(botChannel channel.Channel, cfg *psl.Config, logger *logrus.Logger) {
	if !cfg.Tars.Enabled || botChannel == nil {
		return
	}

	elector := psl.GetLeaderElector()
	if elector == nil {
		return
	}

	elector.OnLeaderChanged(func(isLeader bool) {
		if isLeader {
			logger.Infof("[server] this node became leader, starting channel listener")
			go func() {
				if err := botChannel.StartListening(context.Background()); err != nil {
					logger.Errorf("[server] channel listener error: %v", err)
				}
			}()
		} else {
			logger.Infof("[server] this node lost leadership, stopping channel listener")
			botChannel.StopListening()
		}
	})
}

func shutdown(botChannel channel.Channel) {
	middleware.StopLimiter()
	if elector := psl.GetLeaderElector(); elector != nil {
		elector.Stop()
	}
	if botChannel != nil {
		botChannel.StopListening()
	}
}

func Run(ctx context.Context) error {
	logger := psl.GetLogger()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.POST("/auth/login", sys_user.NewAuthHandler().Login)

	cfg := psl.GetConfig()

	botChannel, err := createChannel(cfg)
	if err != nil {
		logger.WithFields(map[string]interface{}{"error": err.Error()}).Warnf("channel creation failed, tars disabled")
		botChannel = nil
	}

	setupChannel(botChannel, cfg, logger)

	modules := []module.Module{
		health.NewHealthModule(),
		sys_distributed_locker.NewSysDistributedLockModule(),
		sys_server.NewSysServerModule(),
		sys_user.NewSysUserModule(),
		wiki.NewWikiModule(),
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
		shutdown(botChannel)
		return err
	}
}
