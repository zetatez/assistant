package llmproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

var sharedTransport = &http.Transport{
	MaxIdleConns:        8,
	MaxIdleConnsPerHost: 2,
	IdleConnTimeout:     120 * time.Second,
	DisableCompression:  false,
}

type ProviderStatus string

const (
	StatusAvailable   ProviderStatus = "available"
	StatusExhausted   ProviderStatus = "exhausted"
	StatusOffline     ProviderStatus = "offline"
	StatusConfigError ProviderStatus = "config_error"
)

type Strategy string

const (
	StrategyPriority   Strategy = "priority"
	StrategyRandom     Strategy = "random"
	StrategyRoundRobin Strategy = "round_robin"
)

type ProviderConfig struct {
	Name     string   `mapstructure:"name"`
	BaseURL  string   `mapstructure:"base_url"`
	APIKey   string   `mapstructure:"api_key"`
	APIType  string   `mapstructure:"api_type"` // "openai" (default) | "anthropic"
	Models   []string `mapstructure:"models"`
	NeedVPN  bool     `mapstructure:"need_vpn"`
	Priority int      `mapstructure:"priority"`
}

type Config struct {
	ProxiedModel            string           `mapstructure:"proxied_model"`
	VisionModels            []string         `mapstructure:"vision_models"`
	ProbeInterval           int              `mapstructure:"probe_interval"`
	ProxiedAPIKey           string           `mapstructure:"proxied_api_key"`
	Timeout                 int              `mapstructure:"timeout"`
	Temperature             float32          `mapstructure:"temperature"`
	VPN                     string           `mapstructure:"vpn"`
	BalanceStrategy         string           `mapstructure:"balance_strategy"`
	OfflineFailureThreshold int              `mapstructure:"offline_failure_threshold"`
	RateLimitCooldown       int              `mapstructure:"rate_limit_cooldown"`
	Providers               []ProviderConfig `mapstructure:"providers"`
}

type providerState struct {
	ProviderConfig
	status           ProviderStatus
	rateLimitedUntil time.Time
	failures         int
}

type ProxyService struct {
	config       Config
	strategy     Strategy
	rrIndex      int
	mu           sync.RWMutex
	providers    []*providerState
	active       *providerState
	lastModel    string
	httpClient   *http.Client
	vpnClient    *http.Client
	lastActivity time.Time
	probeMu      sync.Mutex
	probeRunning bool
}

func NewProxyService(cfg Config) *ProxyService {
	timeout := 5 * time.Minute
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout) * time.Second
	}

	s := &ProxyService{
		config:   cfg,
		strategy: normalizeStrategy(cfg.BalanceStrategy),
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: sharedTransport,
		},
	}

	if cfg.VPN != "" {
		if dialer, err := newSocksDialer(cfg.VPN); err == nil {
			tr := &http.Transport{
				MaxIdleConns:        8,
				MaxIdleConnsPerHost: 2,
				IdleConnTimeout:     120 * time.Second,
				DisableCompression:  false,
			}
			tr.DialContext = dialer.DialContext
			s.vpnClient = &http.Client{Timeout: timeout, Transport: tr}
		}
	}

	for i, pc := range cfg.Providers {
		if pc.Priority == 0 {
			pc.Priority = i + 1
		}
		s.providers = append(s.providers, &providerState{
			ProviderConfig: pc,
			status:         StatusAvailable,
		})
	}

	sort.SliceStable(s.providers, func(i, j int) bool {
		return s.providers[i].Priority < s.providers[j].Priority
	})
	return s
}

func normalizeStrategy(s string) Strategy {
	switch Strategy(strings.ToLower(strings.TrimSpace(s))) {
	case StrategyRandom:
		return StrategyRandom
	case StrategyRoundRobin:
		return StrategyRoundRobin
	default:
		return StrategyPriority
	}
}

func newSocksDialer(proxyURL string) (interface {
	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
}, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}
	d, err := proxy.FromURL(u, proxy.Direct)
	if err != nil {
		return nil, err
	}
	return &socksDialer{d: d}, nil
}

type socksDialer struct {
	d proxy.Dialer
}

func (s *socksDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return s.d.Dial(network, addr)
}

func (s *ProxyService) Config() Config        { return s.config }
func (s *ProxyService) Strategy() string      { return string(s.strategy) }
func (s *ProxyService) FailureThreshold() int { return s.config.OfflineFailureThreshold }

func (s *ProxyService) HasProviders() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.providers) > 0
}

func (s *ProxyService) ActiveProvider() *ProviderConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.active == nil {
		return nil
	}
	return &s.active.ProviderConfig
}

func (s *ProxyService) PickVisionModel() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, vm := range s.config.VisionModels {
		for _, p := range s.providers {
			if p.status == StatusExhausted || p.status == StatusOffline || p.status == StatusConfigError {
				continue
			}
			if time.Now().Before(p.rateLimitedUntil) {
				continue
			}
			if slices.Contains(p.Models, vm) {
				return vm
			}
		}
	}
	return ""
}

func (s *ProxyService) ProviderStatuses() []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var res []map[string]interface{}
	for _, p := range s.providers {
		res = append(res, map[string]interface{}{
			"name":               p.Name,
			"status":             p.status,
			"priority":           p.Priority,
			"failures":           p.failures,
			"rate_limited_until": p.rateLimitedUntil,
		})
	}
	return res
}

func (s *ProxyService) availableProvidersLocked(model string) []*providerState {
	now := time.Now()
	var out []*providerState
	for _, p := range s.providers {
		if p.status == StatusExhausted || p.status == StatusConfigError {
			continue
		}
		if p.status == StatusOffline {
			continue
		}
		if now.Before(p.rateLimitedUntil) {
			continue
		}
		if model != "" && !slices.Contains(p.Models, model) {
			continue
		}
		out = append(out, p)
	}
	return out
}

func (s *ProxyService) pickByStrategy(candidates []*providerState) *providerState {
	switch s.strategy {
	case StrategyRandom:
		return candidates[rand.Intn(len(candidates))]
	case StrategyRoundRobin:
		p := candidates[s.rrIndex%len(candidates)]
		s.rrIndex++
		return p
	default:
		return candidates[0]
	}
}

func (s *ProxyService) pickBestLocked(model string) *providerState {
	candidates := s.availableProvidersLocked(model)
	if len(candidates) == 0 {
		return nil
	}
	return s.pickByStrategy(candidates)
}

func (s *ProxyService) findNext(current *providerState, modelSpecific bool, originalModel string) *providerState {
	model := ""
	if modelSpecific {
		model = originalModel
	}
	candidates := s.availableProvidersLocked(model)
	var rest []*providerState
	for _, p := range candidates {
		if p != current {
			rest = append(rest, p)
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return s.pickByStrategy(rest)
}

func (s *ProxyService) markProvider(p *providerState, status ProviderStatus, cooldown time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if status != "" {
		p.status = status
	}
	if cooldown > 0 {
		p.rateLimitedUntil = time.Now().Add(cooldown)
	}
	if s.active == p {
		s.active = nil
	}
}

func (s *ProxyService) recordFailure(p *providerState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p.failures++
	threshold := s.config.OfflineFailureThreshold
	if threshold <= 0 {
		threshold = 3
	}
	if p.failures >= threshold && p.status != StatusConfigError {
		p.status = StatusOffline
		if s.active == p {
			s.active = nil
		}
	}
}

func (s *ProxyService) recordSuccess(p *providerState) {
	s.mu.Lock()
	p.failures = 0
	s.mu.Unlock()
}

func (s *ProxyService) classifyProviderFailure(p *providerState, statusCode int) {
	switch {
	case statusCode == http.StatusTooManyRequests:
		cooldown := s.config.RateLimitCooldown
		if cooldown <= 0 {
			cooldown = 60
		}
		s.markProvider(p, "", time.Duration(cooldown)*time.Second)
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden || statusCode == http.StatusNotFound:
		s.markProvider(p, StatusConfigError, 0)
	default:
		s.markProvider(p, StatusExhausted, 0)
	}
}

func (s *ProxyService) ProbeHigherPriority() {
	s.mu.RLock()
	active := s.active
	var candidates []*providerState
	if active == nil {
		for _, p := range s.providers {
			if p.status != StatusAvailable {
				candidates = append(candidates, p)
			}
		}
	} else if len(s.providers) > 0 && active != s.providers[0] {
		for _, p := range s.providers {
			if p == active {
				break
			}
			if p.status != StatusAvailable {
				candidates = append(candidates, p)
			}
		}
	}
	s.mu.RUnlock()

	for _, p := range candidates {
		if p.status == StatusConfigError {
			continue
		}
		client := s.httpClient
		if p.NeedVPN && s.vpnClient != nil {
			client = s.vpnClient
		}
		if probeProvider(p, client) {
			s.mu.Lock()
			p.status = StatusAvailable
			p.rateLimitedUntil = time.Time{}
			p.failures = 0
			s.mu.Unlock()
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	model := s.lastModel
	if s.active == nil {
		if best := s.pickBestLocked(model); best != nil {
			s.active = best
		}
		return
	}
	if best := s.pickBestLocked(model); best != nil && best != s.active {
		s.active = best
	}
}

func probeProvider(p *providerState, client *http.Client) bool {
	if len(p.Models) == 0 {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	probeModel := p.Models[0]
	payload, _ := json.Marshal(map[string]interface{}{
		"model":      probeModel,
		"messages":   []map[string]string{{"role": "user", "content": "hi"}},
		"max_tokens": 1,
		"stream":     false,
	})
	req, err := http.NewRequestWithContext(ctx, "POST",
		strings.TrimRight(p.BaseURL, "/")+"/chat/completions",
		bytes.NewReader(payload))
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "assistant/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return resp.StatusCode == http.StatusOK
}

func (s *ProxyService) NotifyActivity() {
	s.probeMu.Lock()
	s.lastActivity = time.Now()
	if !s.probeRunning {
		s.probeRunning = true
		s.probeMu.Unlock()
		go s.probeLoop()
	} else {
		s.probeMu.Unlock()
	}
}

func (s *ProxyService) probeLoop() {
	interval := 30 * time.Second
	if s.config.ProbeInterval > 0 {
		interval = time.Duration(s.config.ProbeInterval) * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		s.ProbeHigherPriority()

		s.probeMu.Lock()
		if time.Since(s.lastActivity) > interval*2 {
			s.probeRunning = false
			s.probeMu.Unlock()
			return
		}
		s.probeMu.Unlock()
	}
}

func (s *ProxyService) ForwardChat(ctx context.Context, cfg ProviderConfig, bodyReader io.ReadCloser) (*http.Response, error) {
	raw, err := io.ReadAll(bodyReader)
	bodyReader.Close()
	if err != nil {
		return nil, err
	}

	isAnthropic := strings.EqualFold(cfg.APIType, "anthropic")
	targetURL := strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"
	body := raw
	if isAnthropic {
		targetURL = strings.TrimRight(cfg.BaseURL, "/") + "/v1/messages"
		body, err = OpenAIReqToAnthropic(raw)
		if err != nil {
			return nil, fmt.Errorf("convert to anthropic request: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "assistant/1.0")

	client := s.httpClient
	if cfg.NeedVPN && s.vpnClient != nil {
		client = s.vpnClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if isAnthropic && resp.StatusCode < 400 {
		return convertAnthropicResponse(resp)
	}
	return resp, nil
}

func (s *ProxyService) Forward(ctx context.Context, reqMap map[string]interface{}, requestedModel string) (*http.Response, error) {
	var p *providerState
	var modelSpecific bool

	s.mu.Lock()
	if requestedModel == s.config.ProxiedModel || requestedModel == "" {
		s.lastModel = ""
		p = s.pickBestLocked("")
	} else {
		s.lastModel = requestedModel
		p = s.pickBestLocked(requestedModel)
		if p == nil {
			p = s.pickBestLocked("")
		} else {
			modelSpecific = true
		}
	}
	if p != nil {
		s.active = p
	}
	s.mu.Unlock()

	if p == nil {
		s.NotifyActivity()
		return nil, s.noProviderError()
	}

	for {
		reqMap["model"] = resolveModel(p, requestedModel)
		stripKnownBad(reqMap)
		body, _ := json.Marshal(reqMap)

		resp, err := s.ForwardChat(ctx, p.ProviderConfig, io.NopCloser(bytes.NewReader(body)))
		if err != nil {
			s.recordFailure(p)
			next := s.findNext(p, modelSpecific, requestedModel)
			if next == nil {
				s.NotifyActivity()
				return nil, fmt.Errorf("all providers failed: %s: %v", p.Name, err)
			}
			p = next
			continue
		}

		if resp.StatusCode == http.StatusOK {
			s.recordSuccess(p)
			s.NotifyActivity()
			return resp, nil
		}

		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		resp.Body.Close()

		s.classifyProviderFailure(p, resp.StatusCode)

		next := s.findNext(p, modelSpecific, requestedModel)
		if next == nil {
			s.NotifyActivity()
			errMsg := providerErrorMessage(bodyBytes)
			return nil, &HTTPError{Code: resp.StatusCode, Message: errMsg}
		}
		p = next
	}
}

type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string { return e.Message }

var knownBadFields = []string{"promptCacheKey"}

func stripKnownBad(req map[string]interface{}) {
	for _, k := range knownBadFields {
		delete(req, k)
	}
}

func resolveModel(p *providerState, requested string) string {
	if len(p.Models) == 0 {
		return requested
	}
	if requested != "" && slices.Contains(p.Models, requested) {
		return requested
	}
	return p.Models[0]
}

func (s *ProxyService) noProviderError() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	exhausted, offline, cfgErr := 0, 0, 0
	for _, p := range s.providers {
		switch p.status {
		case StatusExhausted:
			exhausted++
		case StatusOffline:
			offline++
		case StatusConfigError:
			cfgErr++
		}
	}
	if exhausted > 0 || cfgErr > 0 {
		return fmt.Errorf("no available provider (%d exhausted, %d offline, %d config_error)", exhausted, offline, cfgErr)
	}
	return fmt.Errorf("no available provider (%d offline)", offline)
}

func providerErrorMessage(body []byte) string {
	if len(body) == 0 {
		return "provider failed"
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return string(body)
	}
	if msg, ok := parsed["error"].(string); ok {
		return msg
	}
	if errObj, ok := parsed["error"].(map[string]interface{}); ok {
		if msg, ok := errObj["message"].(string); ok {
			return msg
		}
	}
	return string(body)
}
