package wiki

import (
	"context"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type IndexManager struct {
	cfg          Config
	index        map[string]*IndexedEntry
	indexMu      sync.RWMutex
	contentCache map[string]*cachedFile
	cacheMu      sync.RWMutex
	stopCh       chan struct{}
	stoppedCh    chan struct{}
	reranker     Reranker
}

type cachedFile struct {
	content string
	modTime time.Time
	expiry  time.Time
}

const contentCacheTTL = 30 * time.Second

type IndexedEntry struct {
	Path  string
	Title string
}

func NewIndexManager(cfg Config) *IndexManager {
	m := &IndexManager{
		cfg:          cfg,
		index:        make(map[string]*IndexedEntry),
		contentCache: make(map[string]*cachedFile),
	}
	if cfg.Enabled && cfg.Dir != "" {
		m.BuildIndex()
	}
	return m
}

func (m *IndexManager) BuildIndex() {
	m.indexMu.Lock()
	defer m.indexMu.Unlock()

	m.contentCache = make(map[string]*cachedFile)

	dir := expandDir(m.cfg.Dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	m.index = make(map[string]*IndexedEntry)

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != ".git" && entry.Name() != "raw" {
			m.scanDir(filepath.Join(dir, entry.Name()))
		}
	}
}

func (m *IndexManager) scanDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != ".git" && entry.Name() != "raw" {
			m.scanDir(filepath.Join(dir, entry.Name()))
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".md") || entry.Name() == "index.md" || entry.Name() == "log.md" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		e := m.parseFile(path)
		m.index[path] = e
	}
}

func (m *IndexManager) parseFile(path string) *IndexedEntry {
	content, err := os.ReadFile(path)
	if err != nil {
		return &IndexedEntry{Path: path}
	}

	title := extractTitle(string(content), filepath.Base(path))

	return &IndexedEntry{
		Path:  path,
		Title: title,
	}
}

func extractTitle(content, filename string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	re := regexp.MustCompile(`(?i)^<!--\s*title:\s*(.+?)\s*-->`)
	maxLines := 10
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for _, line := range lines[:maxLines] {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

func (m *IndexManager) GrepContent(query string, limit int) []*GrepHit {
	m.indexMu.RLock()
	defer m.indexMu.RUnlock()

	queryLower := strings.ToLower(query)
	var results []*GrepHit

	for _, e := range m.index {
		contentStr := m.getFileContent(e.Path)
		if contentStr == "" {
			continue
		}
		contentLower := strings.ToLower(contentStr)

		idx := strings.Index(contentLower, queryLower)
		if idx == -1 {
			continue
		}

		start := idx
		end := idx + len(query)

		startMargin := 300
		endMargin := 400

		for startMargin > 0 && start > 0 && contentStr[start-1] != '\n' {
			start--
			startMargin--
		}
		for startMargin > 0 && start > 0 && contentStr[start-1] != '.' && contentStr[start-1] != '!' && contentStr[start-1] != '?' {
			start--
			startMargin--
		}

		for endMargin > 0 && end < len(contentStr) && contentStr[end] != '\n' {
			end++
			endMargin--
		}
		for endMargin > 0 && end < len(contentStr) && contentStr[end] != '.' && contentStr[end] != '!' && contentStr[end] != '?' {
			end++
			endMargin--
		}

		if start < 0 {
			start = 0
		}
		if end > len(contentStr) {
			end = len(contentStr)
		}

		snippet := strings.TrimSpace(contentStr[start:end])
		trimmed := strings.TrimLeft(snippet, " \t")
		if len(snippet) != len(trimmed) || start > 0 {
			snippet = "..." + trimmed
		}
		if end < len(contentStr) {
			snippet = snippet + "..."
		}

		score := 10.0
		if strings.Contains(strings.ToLower(e.Title), queryLower) {
			score += 5
		}

		results = append(results, &GrepHit{
			Entry:   e,
			Snippet: snippet,
			Score:   int(score),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

func (m *IndexManager) getFileContent(path string) string {
	now := time.Now()
	m.cacheMu.RLock()
	if c, ok := m.contentCache[path]; ok && now.Before(c.expiry) {
		m.cacheMu.RUnlock()
		return c.content
	}
	m.cacheMu.RUnlock()

	info, err := os.Stat(path)
	if err != nil {
		return ""
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	contentStr := string(content)

	m.cacheMu.Lock()
	if c, ok := m.contentCache[path]; ok && c.modTime.Equal(info.ModTime()) {
		c.content = contentStr
		c.expiry = now.Add(contentCacheTTL)
	} else {
		m.contentCache[path] = &cachedFile{
			content: contentStr,
			modTime: info.ModTime(),
			expiry:  now.Add(contentCacheTTL),
		}
	}
	m.cacheMu.Unlock()
	return contentStr
}

type GrepHit struct {
	Entry   *IndexedEntry
	Snippet string
	Score   int
}

func (m *IndexManager) SetReranker(r Reranker) {
	m.reranker = r
}

func (m *IndexManager) Search(ctx context.Context, query string, limit int) ([]*RerankResult, error) {
	grepResults := m.GrepContent(query, limit)
	if len(grepResults) == 0 {
		return nil, nil
	}
	if m.reranker == nil {
		out := make([]*RerankResult, len(grepResults))
		for i, r := range grepResults {
			out[i] = &RerankResult{Entry: r.Entry, Snippet: r.Snippet, Score: float64(r.Score), Reason: ""}
		}
		return out, nil
	}
	return m.reranker.Rerank(ctx, query, grepResults)
}

func (m *IndexManager) Count() int {
	m.indexMu.RLock()
	defer m.indexMu.RUnlock()
	return len(m.index)
}

func (m *IndexManager) StartBackgroundRefresh(interval time.Duration) {
	if m.stopCh != nil {
		return
	}
	m.stopCh = make(chan struct{})
	m.stoppedCh = make(chan struct{})

	go func() {
		defer close(m.stoppedCh)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.BuildIndex()
			case <-m.stopCh:
				return
			}
		}
	}()
}

func (m *IndexManager) Stop() {
	if m.stopCh == nil {
		return
	}
	close(m.stopCh)
	<-m.stoppedCh
	m.stopCh = nil
	m.stoppedCh = nil
}

func (m *IndexManager) RefreshIndex() {
	m.BuildIndex()
}

func expandDir(dir string) string {
	if strings.HasPrefix(dir, "~/") {
		usr, err := user.Current()
		if err == nil {
			dir = filepath.Join(usr.HomeDir, dir[2:])
		}
	}
	if !filepath.IsAbs(dir) {
		if absDir, err := filepath.Abs(dir); err == nil {
			return absDir
		}
	}
	return dir
}
