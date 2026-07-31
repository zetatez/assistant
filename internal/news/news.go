package news

import (
	"context"
	"time"

	"assistant/pkg/dwmblocknotify"
	pkgnews "assistant/pkg/news"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Enabled        bool   `mapstructure:"enabled"`
	Interval       string `mapstructure:"interval"`
	NotifyInterval string `mapstructure:"notify_interval"`
	NotifyTTL      string `mapstructure:"notify_ttl"`
}

type Service struct {
	cfg            Config
	collector      *pkgnews.Collector
	logger         *logrus.Logger
	fetchInterval  time.Duration
	notifyTTL      time.Duration
	notifyInterval time.Duration
}

func NewService(cfg Config, logger *logrus.Logger) *Service {
	s := &Service{
		cfg:       cfg,
		collector: pkgnews.New(),
		logger:    logger,
	}
	s.fetchInterval = parseDuration(cfg.Interval, 30*time.Minute)
	s.notifyTTL = parseDuration(cfg.NotifyTTL, 3*time.Second)
	s.notifyInterval = parseDuration(cfg.NotifyInterval, 16*time.Second)
	return s
}

func (s *Service) Start(ctx context.Context) {
	if !s.cfg.Enabled {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"interval":        s.fetchInterval,
		"notify_interval": s.notifyInterval,
		"notify_ttl":      s.notifyTTL,
	}).Info("news: service started")
	go s.loop(ctx)
}

func (s *Service) loop(ctx context.Context) {
	items := s.fetch(ctx)

	fetchTicker := time.NewTicker(s.fetchInterval)
	defer fetchTicker.Stop()
	notifyTicker := time.NewTicker(s.notifyInterval)
	defer notifyTicker.Stop()

	idx := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-fetchTicker.C:
			items = s.fetch(ctx)
			idx = 0
		case <-notifyTicker.C:
			if len(items) == 0 {
				continue
			}
			dwmblocknotify.POST(items[idx].Title, s.notifyTTL)
			idx = (idx + 1) % len(items)
		}
	}
}

func (s *Service) fetch(ctx context.Context) []pkgnews.Item {
	var all []pkgnews.Item
	for _, p := range s.collector.Providers() {
		items, err := s.collector.Fetch(ctx, p, 16)
		if err != nil {
			s.logger.WithError(err).Warnf("news: fetch %s failed", p)
			continue
		}
		all = append(all, items...)
	}
	s.logger.Infof("news: fetched %d items", len(all))
	return all
}

func parseDuration(s string, fallback time.Duration) time.Duration {
	if s == "" {
		return fallback
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}
