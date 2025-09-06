package tars

import (
	"assistant/internal/app/module"
	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/channel"
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

type Module struct {
	name    string
	handler *Handler
}

func NewModule(ctx context.Context, ch channel.Channel) module.Module {
	cfg := psl.GetConfig()
	if !cfg.Tars.Enabled {
		return &Module{name: ""}
	}

	db := psl.GetDB()
	if db == nil {
		psl.GetLogger().Errorf("tars: database not available")
		return &Module{name: ""}
	}

	queries := repo.New(db)
	memory := NewMemoryService(queries, psl.GetLogger())
	wikiRepo := repo.NewWikiRepo(repo.New(db))
	handler := NewHandler(ch, memory, wikiRepo, psl.GetLogger())
	handler.Register()

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				if err := memory.CleanupOld(cleanupCtx); err != nil {
					psl.GetLogger().Errorf("tars cleanup error: %v", err)
				}
				cancel()
			case <-ctx.Done():
				return
			}
		}
	}()

	return &Module{name: "tars", handler: handler}
}

func (m *Module) Name() string { return m.name }

func (m *Module) Register(r *gin.RouterGroup) {
}

func (m *Module) Middleware() []gin.HandlerFunc {
	return nil
}
