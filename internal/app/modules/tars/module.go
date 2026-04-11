package tars

import (
	"context"
	"sync"
	"time"

	"assistant/internal/app/module"
	"assistant/internal/app/modules/tars/knowledge"
	"assistant/internal/app/modules/tars/memory"
	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/channel"
	"assistant/pkg/llm"
	"assistant/pkg/wiki"

	"github.com/gin-gonic/gin"
)

type Module struct {
	name        string
	handler     *Handler
	wikiManager *wiki.IndexManager
	done        chan struct{}
	stopOnce    sync.Once
	wg          sync.WaitGroup
}

func NewModule(ctx context.Context, ch channel.Channel) module.Module {
	cfg := psl.GetConfig()
	if !cfg.Tars.Enabled {
		return &Module{}
	}

	db := psl.GetDB()
	if db == nil {
		psl.GetLogger().Errorf("tars: database not available")
		return &Module{}
	}

	llmCfg := cfg.LLM
	var llmClient llm.Client
	if llmCfg.Provider != "" {
		var err error
		llmClient, err = llm.NewClient(llmCfg.Provider, llm.Config{
			APIKey:     llmCfg.APIKey,
			BaseURL:    llmCfg.BaseURL,
			Model:      llmCfg.Model,
			Timeout:    llmCfg.Timeout,
			MaxRetries: 3,
		})
		if err != nil {
			psl.GetLogger().Warnf("tars: failed to create LLM client: %v", err)
		}
	}

	queries := repo.New(db)
	memory := memory.NewMemoryService(queries, llmClient, psl.GetLogger(), &cfg.Tars, llmCfg.Model)
	knowledgeManager := knowledge.NewManager(queries, db, llmClient, llmCfg.Model, psl.GetLogger())
	sessionManager := knowledge.NewSessionManager(queries, llmClient, llmCfg.Model, psl.GetLogger())
	wikiManager := wiki.NewIndexManager(wiki.Config{
		Enabled: cfg.Tars.Wiki.Enabled,
		Dir:     cfg.Tars.Wiki.Dir,
	})
	if cfg.Tars.Wiki.Enabled && llmClient != nil {
		wikiManager.SetReranker(wiki.NewLLMReranker(llmClient, llmCfg.Model))
	}
	handler := NewHandler(ch, memory, llmClient, knowledgeManager, sessionManager, wikiManager, psl.GetLogger(), llmCfg.Model, cfg.Tars.LLMTemperature)
	handler.Register()

	done := make(chan struct{})
	m := &Module{name: "tars", handler: handler, wikiManager: wikiManager, done: done}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		sessionMaxAge := 7 * 24 * time.Hour
		for {
			select {
			case <-ticker.C:
				cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				if err := memory.CleanupOld(cleanupCtx); err != nil {
					psl.GetLogger().Errorf("tars cleanup error: %v", err)
				}
				olderThan := time.Now().AddDate(0, 0, -90)
				if err := knowledgeManager.CleanupOldKnowledge(cleanupCtx, olderThan); err != nil {
					psl.GetLogger().Errorf("tars knowledge cleanup error: %v", err)
				}
				cleaned := memory.CleanupShortTermSessions(sessionMaxAge)
				if cleaned > 0 {
					psl.GetLogger().Infof("tars: cleaned %d stale short-term sessions", cleaned)
				}
				cancel()
			case <-done:
				cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				memory.CleanupOld(cleanupCtx)
				olderThan := time.Now().AddDate(0, 0, -90)
				knowledgeManager.CleanupOldKnowledge(cleanupCtx, olderThan)
				cancel()
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return m
}

func (m *Module) Name() string { return m.name }

func (m *Module) Register(r *gin.RouterGroup) {
}

func (m *Module) Middleware() []gin.HandlerFunc {
	return nil
}

func (m *Module) Stop() {
	m.stopOnce.Do(func() {
		if m.done != nil {
			close(m.done)
		}
	})
	if m.handler != nil {
		m.handler.Stop()
	}
	if m.wikiManager != nil {
		m.wikiManager.Stop()
	}
	m.wg.Wait()
}
