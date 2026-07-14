package tars

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"assistant/pkg/llm"
)

const (
	webSearchTimeout = 10 * time.Second
	maxWebResults    = 5
	maxWikiResults   = 5
	maxSnippetLen    = 500
	maxReadLen       = 8000
)

type ToolExecutor struct {
	wikiManager *IndexManager
	logger      Logger
}

func NewToolExecutor(wikiMgr *IndexManager, logger Logger) *ToolExecutor {
	return &ToolExecutor{
		wikiManager: wikiMgr,
		logger:      logger,
	}
}

func (e *ToolExecutor) Definitions() []llm.ToolDefinition {
	defs := []llm.ToolDefinition{
		{
			Name:        "grep_wiki",
			Description: "Search the local wiki/knowledge base for relevant information. Use this when the user asks about topics that might be documented locally.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Search keywords to look up in the local knowledge base",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "read_wiki",
			Description: "Read the full content of a wiki document by its path. Use this after grep_wiki to get more context when the snippet is not enough.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "The file path returned by grep_wiki",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "web_search",
			Description: "Search the internet for current information. Use this when you lack knowledge to answer the user's question and need to look up information online.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Search query for the web search",
					},
				},
				"required": []string{"query"},
			},
		},
	}
	return defs
}

func (e *ToolExecutor) Execute(ctx context.Context, name, argsJSON string) (string, error) {
	switch name {
	case "grep_wiki":
		return e.execGrepWiki(ctx, argsJSON)
	case "read_wiki":
		return e.execReadWiki(ctx, argsJSON)
	case "web_search":
		return e.execWebSearch(ctx, argsJSON)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (e *ToolExecutor) execGrepWiki(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	if args.Query == "" {
		return "No results found.", nil
	}
	if e.wikiManager == nil {
		return "Wiki knowledge base is not available.", nil
	}

	hits := e.wikiManager.GrepContent(args.Query, maxWikiResults)
	if len(hits) == 0 {
		return "No results found in local knowledge base.", nil
	}

	var sb strings.Builder
	for i, hit := range hits {
		if i > 0 {
			sb.WriteString("\n---\n")
		}
		title := hit.Entry.Title
		if title == "" {
			title = hit.Entry.Path
		}
		sb.WriteString(fmt.Sprintf("**%s** (path: %s)\n%s", title, hit.Entry.Path, hit.Snippet))
		if len(sb.String()) > maxSnippetLen*maxWikiResults {
			break
		}
	}
	return sb.String(), nil
}

func (e *ToolExecutor) execReadWiki(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	if args.Path == "" {
		return "", fmt.Errorf("path is required")
	}
	if e.wikiManager == nil {
		return "Wiki knowledge base is not available.", nil
	}

	entry := e.wikiManager.FindEntry(args.Path)
	if entry == nil {
		return "File not found in wiki index.", nil
	}

	data, err := os.ReadFile(entry.Path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	content := string(data)
	if len(content) > maxReadLen {
		content = content[:maxReadLen] + "\n... [truncated]"
	}
	return content, nil
}

func (e *ToolExecutor) execWebSearch(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	if args.Query == "" {
		return "No search results.", nil
	}

	searchCtx, cancel := context.WithTimeout(ctx, webSearchTimeout)
	defer cancel()

	results, err := duckDuckGoSearch(searchCtx, args.Query)
	if err != nil {
		e.logger.Warnf("tars: web_search error: %v", err)
		return fmt.Sprintf("Web search failed: %v", err), nil
	}
	if len(results) == 0 {
		return "No web search results found.", nil
	}

	var sb strings.Builder
	for i, r := range results {
		if i >= maxWebResults {
			break
		}
		if i > 0 {
			sb.WriteString("\n---\n")
		}
		sb.WriteString(fmt.Sprintf("**%s**\n%s", r.Title, r.Snippet))
	}
	return sb.String(), nil
}

type searchResult struct {
	Title   string
	Snippet string
}

func duckDuckGoSearch(ctx context.Context, query string) ([]searchResult, error) {
	reqURL := "https://lite.duckduckgo.com/lite/?q=" + url.QueryEscape(query) + "&kl=cn-zh"

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, err
	}

	return parseDuckDuckGoLite(string(body)), nil
}

var ddgTitleRe = regexp.MustCompile(`<a[^>]*class="result-link"[^>]*>(.*?)</a>`)
var ddgSnippetRe = regexp.MustCompile(`<td[^>]*class="result-snippet"[^>]*>(.*?)</td>`)
var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

func parseDuckDuckGoLite(html string) []searchResult {
	titles := ddgTitleRe.FindAllStringSubmatch(html, -1)
	snippets := ddgSnippetRe.FindAllStringSubmatch(html, -1)

	var results []searchResult
	for i, t := range titles {
		title := cleanHTML(t[1])
		if title == "" {
			continue
		}
		snippet := ""
		if i < len(snippets) {
			snippet = cleanHTML(snippets[i][1])
		}
		if len(snippet) > maxSnippetLen {
			snippet = snippet[:maxSnippetLen] + "..."
		}
		results = append(results, searchResult{Title: title, Snippet: snippet})
	}
	return results
}

func cleanHTML(s string) string {
	return strings.TrimSpace(html.UnescapeString(htmlTagRe.ReplaceAllString(s, "")))
}

type Config struct {
	Enabled bool
	Dir     string
}

type IndexedEntry struct {
	Path  string
	Title string
}

type GrepHit struct {
	Entry   *IndexedEntry
	Snippet string
	Score   int
}

type IndexManager struct {
	cfg       Config
	index     map[string]*IndexedEntry
	indexMu   sync.RWMutex
	stopCh    chan struct{}
	stoppedCh chan struct{}
}

func NewIndexManager(cfg Config) *IndexManager {
	m := &IndexManager{
		cfg:   cfg,
		index: make(map[string]*IndexedEntry),
	}
	if cfg.Enabled && cfg.Dir != "" {
		m.buildIndex()
	}
	return m
}

func (m *IndexManager) buildIndex() {
	m.indexMu.Lock()
	defer m.indexMu.Unlock()
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
		m.index[path] = m.parseFile(path)
	}
}

func (m *IndexManager) parseFile(path string) *IndexedEntry {
	content, err := os.ReadFile(path)
	if err != nil {
		return &IndexedEntry{Path: path}
	}
	title := extractTitle(string(content), filepath.Base(path))
	return &IndexedEntry{Path: path, Title: title}
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
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

func (m *IndexManager) FindEntry(path string) *IndexedEntry {
	m.indexMu.RLock()
	defer m.indexMu.RUnlock()
	if e, ok := m.index[path]; ok {
		return e
	}
	for k, e := range m.index {
		if strings.HasSuffix(k, path) || filepath.Base(k) == path {
			return e
		}
	}
	return nil
}

func (m *IndexManager) GrepContent(query string, limit int) []*GrepHit {
	m.indexMu.RLock()
	defer m.indexMu.RUnlock()

	queryLower := strings.ToLower(query)
	var results []*GrepHit

	for _, e := range m.index {
		data, err := os.ReadFile(e.Path)
		if err != nil {
			continue
		}
		contentStr := string(data)
		contentLower := strings.ToLower(contentStr)

		idx := strings.Index(contentLower, queryLower)
		if idx == -1 {
			continue
		}

		start := idx
		end := idx + len(query)

		for startMargin := 300; startMargin > 0 && start > 0 && contentStr[start-1] != '\n'; start-- {
			startMargin--
		}
		for startMargin := 300; startMargin > 0 && start > 0 && contentStr[start-1] != '.' && contentStr[start-1] != '!' && contentStr[start-1] != '?'; start-- {
			startMargin--
		}
		for endMargin := 400; endMargin > 0 && end < len(contentStr) && contentStr[end] != '\n'; end++ {
			endMargin--
		}
		for endMargin := 400; endMargin > 0 && end < len(contentStr) && contentStr[end] != '.' && contentStr[end] != '!' && contentStr[end] != '?'; end++ {
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

		score := 10
		if strings.Contains(strings.ToLower(e.Title), queryLower) {
			score += 5
		}

		results = append(results, &GrepHit{
			Entry:   e,
			Snippet: snippet,
			Score:   score,
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
				m.buildIndex()
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

func expandDir(dir string) string {
	if strings.HasPrefix(dir, "~/") {
		if usr, err := user.Current(); err == nil {
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
