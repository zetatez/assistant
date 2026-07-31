package news

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"assistant/pkg/dwmblocknotify"
	"assistant/pkg/news_collector"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

type Config struct {
	Enabled        bool   `mapstructure:"enabled"`
	Interval       string `mapstructure:"interval"`
	NotifyInterval string `mapstructure:"notify_interval"`
	NotifyTTL      string `mapstructure:"notify_ttl"`
}

type Service struct {
	enabled        bool
	collector      *news_collector.Collector
	logger         *logrus.Logger
	fetchInterval  time.Duration
	notifyTTL      time.Duration
	notifyInterval time.Duration
}

func NewService(cfg Config, vpn string, logger *logrus.Logger) *Service {
	collector := news_collector.New()
	if vpn != "" {
		if u, err := url.Parse(vpn); err == nil {
			if d, err := proxy.FromURL(u, proxy.Direct); err == nil {
				collector.SetClient(&http.Client{
					Timeout: 8 * time.Second,
					Transport: &http.Transport{
						DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
							return d.Dial(network, addr)
						},
					},
				})
				logger.Infof("news: using VPN %s", vpn)
			}
		}
	}
	return &Service{
		enabled:        cfg.Enabled,
		collector:      collector,
		logger:         logger,
		fetchInterval:  parseDuration(cfg.Interval, 30*time.Minute),
		notifyTTL:      parseDuration(cfg.NotifyTTL, 3*time.Second),
		notifyInterval: parseDuration(cfg.NotifyInterval, 16*time.Second),
	}
}

func (s *Service) Start(ctx context.Context) {
	if !s.enabled {
		return
	}
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

func (s *Service) fetch(ctx context.Context) []news_collector.Item {
	var all []news_collector.Item
	for _, p := range s.collector.Providers() {
		if items, err := s.collector.Fetch(ctx, p, 16); err == nil {
			all = append(all, items...)
		}
	}
	return all
}

func parseDuration(s string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	return fallback
}
