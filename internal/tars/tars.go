package tars

import (
	"context"
	"time"

	"assistant/pkg/channel"
	"assistant/pkg/llm"
)

type TarsConfig struct {
	Enabled          bool
	DataDir          string
	Temperature      float32
	MaxHistoryMB     int
	SummarizeEvery   int
	MaxShortTerm     int
	WikiEnabled      bool
	WikiDir          string
	WebSearchEnabled bool
}

type Service struct {
	handler     *Handler
	wikiManager *IndexManager
	ch          channel.Channel
	logger      Logger
}

func NewService(ch channel.Channel, llmClient llm.Client, llmModel string, cfg *TarsConfig, logger Logger) *Service {
	if !cfg.Enabled {
		return &Service{logger: logger}
	}

	if llmClient == nil {
		logger.Errorf("tars: LLM client is nil")
		return &Service{logger: logger}
	}

	memory := NewMemoryService(expandDir(cfg.DataDir), llmClient, llmModel, logger)

	if cfg.MaxShortTerm > 0 {
		defaultMaxHistory = cfg.MaxShortTerm
	}
	if cfg.SummarizeEvery > 0 {
		summarizeThreshold = cfg.SummarizeEvery
	}
	if cfg.MaxHistoryMB > 0 {
		maxHistoryBytes = int64(cfg.MaxHistoryMB) * 1024 * 1024
	}
	llmTemperature := float32(0.7)
	if cfg.Temperature > 0 {
		llmTemperature = cfg.Temperature
	}

	wikiManager := NewIndexManager(Config{
		Enabled: cfg.WikiEnabled,
		Dir:     cfg.WikiDir,
	})

	handler := NewHandler(ch, memory, llmClient, wikiManager, logger, llmModel, llmTemperature, 120*time.Second)
	handler.Register()

	return &Service{
		handler:     handler,
		wikiManager: wikiManager,
		ch:          ch,
		logger:      logger,
	}
}

func (s *Service) Start(ctx context.Context) error {
	if s.ch == nil {
		return nil
	}
	go func() {
		if err := s.ch.StartListening(ctx); err != nil {
			s.logger.Errorf("tars: channel listening error: %v", err)
		}
	}()

	if s.wikiManager != nil {
		s.wikiManager.StartBackgroundRefresh(10 * time.Minute)
	}

	go s.cleanupLoop(ctx)
	return nil
}

func (s *Service) Stop() {
	if s.wikiManager != nil {
		s.wikiManager.Stop()
	}
	if s.ch != nil {
		s.ch.StopListening()
	}
}

func (s *Service) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	sessionMaxAge := 7 * 24 * time.Hour
	for {
		select {
		case <-ticker.C:
			if s.handler != nil && s.handler.memory != nil {
				cleaned := s.handler.memory.shortTerm.CleanupOldSessions(sessionMaxAge)
				if cleaned > 0 {
					s.logger.Infof("tars: cleaned %d stale short-term sessions", cleaned)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
