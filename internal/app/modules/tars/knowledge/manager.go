package knowledge

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"assistant/internal/app/repo"
	"assistant/pkg/llm"
)

type Manager struct {
	repo       *repo.Queries
	db         *sql.DB
	extractor  *Extractor
	integrator *Integrator
	logger     Logger
}

func NewManager(repo *repo.Queries, db *sql.DB, llmClient llm.Client, llmModel string, logger Logger) *Manager {
	return &Manager{
		repo:       repo,
		db:         db,
		extractor:  NewExtractor(llmClient, llmModel),
		integrator: NewIntegrator(repo, db, llmClient, llmModel, logger),
		logger:     logger,
	}
}

func (m *Manager) IntegrateMessage(ctx context.Context, sessionID, userMsg, assistantMsg string) error {
	messages := []string{userMsg, assistantMsg}

	entities, err := m.extractor.ExtractEntities(ctx, messages)
	if err != nil {
		m.logger.Warnf("extract entities error: %v", err)
		return err
	}
	if len(entities) == 0 {
		return nil
	}

	for i := range entities {
		entities[i].SessionID = sessionID
	}

	var relations []Relation
	if len(entities) >= 2 {
		relations, err = m.extractor.ExtractRelations(ctx, entities, nil, userMsg+" "+assistantMsg)
		if err != nil {
			m.logger.Warnf("extract relations error: %v", err)
		}
		for i := range relations {
			relations[i].SessionID = sessionID
		}
	}

	tx, err := m.integrator.DB().BeginTx(ctx, nil)
	if err != nil {
		m.logger.Errorf("begin transaction error: %v", err)
		return err
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	txQueries := m.integrator.WithTx(tx)
	nameToID, err := m.integrator.SaveEntitiesAndGetIDsTx(ctx, entities, txQueries)
	if err != nil {
		m.logger.Errorf("save entities error: %v", err)
		return err
	}

	for i := range relations {
		if err := m.integrator.SaveRelation(ctx, &relations[i], tx); err != nil {
			m.logger.Errorf("save relation error: %v", err)
			return err
		}
	}

	for _, entity := range entities {
		entity.ID = nameToID[entity.Name]
		content, err := m.integrator.GenerateKnowledgeContent(ctx, &entity, messages)
		if err != nil {
			m.logger.Warnf("generate knowledge content error: %v", err)
			continue
		}
		page := &KnowledgePage{
			SessionID:      sessionID,
			EntityID:       entity.ID,
			Title:          entity.Name,
			Content:        content,
			SourceMessages: userMsg + "|" + assistantMsg,
		}
		if err := m.integrator.CreateKnowledgePage(ctx, page, tx); err != nil {
			m.logger.Errorf("create knowledge page error: %v", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		m.logger.Errorf("commit transaction error: %v", err)
		return err
	}
	committed = true

	logDetail, _ := json.Marshal(map[string]interface{}{
		"entities_count": len(entities),
		"user_msg_len":   len(userMsg),
	})
	if err := m.integrator.AppendLog(ctx, sessionID, ActionIngest, string(logDetail)); err != nil {
		m.logger.Warnf("append log error: %v", err)
	}

	return nil
}

func (m *Manager) GetRelatedKnowledge(ctx context.Context, sessionID string, query string) (string, error) {
	keywords := extractKeywordsFromQuery(query)
	if len(keywords) == 0 {
		return "", nil
	}
	pages, err := m.integrator.SearchRelatedKnowledge(ctx, sessionID, keywords)
	if err != nil {
		return "", err
	}
	return m.integrator.BuildWikiContext(pages), nil
}

type LintResult struct {
	Contradictions []string
	Orphans        []string
	Suggestions    []string
}

func (m *Manager) LintKnowledge(ctx context.Context, sessionID string) (*LintResult, error) {
	result := &LintResult{}

	entities, err := m.repo.GetChatEntities(ctx, repo.GetChatEntitiesParams{
		SessionID: sessionID,
		Limit:     100,
		Offset:    0,
	})
	if err != nil {
		return nil, err
	}

	entityMap := make(map[int64]string)
	for _, e := range entities {
		if e.ID > 0 {
			entityMap[e.ID] = e.EntityName
		}
	}

	relations, err := m.repo.GetChatRelations(ctx, sessionID)
	if err != nil {
		m.logger.Warnf("get relations error: %v", err)
	}

	for _, rel := range relations {
		if rel.RelationType == "contradicts" {
			fromName := entityMap[rel.FromEntityID]
			toName := entityMap[rel.ToEntityID]
			if fromName != "" && toName != "" {
				result.Contradictions = append(result.Contradictions, fromName+" contradicts "+toName)
			}
		}
	}

	var orphanedCount int
	for _, e := range entities {
		hasRelation := false
		for _, rel := range relations {
			if rel.FromEntityID == e.ID || rel.ToEntityID == e.ID {
				hasRelation = true
				break
			}
		}
		if !hasRelation {
			orphanedCount++
		}
	}
	if orphanedCount > 0 {
		result.Orphans = append(result.Orphans, fmt.Sprintf("%d orphaned entities with no connections", orphanedCount))
	}

	result.Suggestions = append(result.Suggestions, "Regularly check if new messages need integration into existing knowledge pages")

	return result, nil
}

type IndexEntry struct {
	Type      string
	Name      string
	Summary   string
	UpdatedAt time.Time
}

func (m *Manager) BuildIndex(ctx context.Context, sessionID string) ([]IndexEntry, error) {
	entities, err := m.repo.GetChatEntities(ctx, repo.GetChatEntitiesParams{
		SessionID: sessionID,
		Limit:     100,
		Offset:    0,
	})
	if err != nil {
		return nil, err
	}

	var entries []IndexEntry
	for _, e := range entities {
		desc := ""
		if e.Description.Valid {
			desc = e.Description.String
		}
		updatedAt := time.Now()
		if e.UpdatedAt.Valid {
			updatedAt = e.UpdatedAt.Time
		}
		entries = append(entries, IndexEntry{
			Type:      e.EntityType,
			Name:      e.EntityName,
			Summary:   desc,
			UpdatedAt: updatedAt,
		})
	}

	return entries, nil
}

func (m *Manager) AppendLog(ctx context.Context, sessionID, action, detail string) error {
	_, err := m.repo.CreateChatLog(ctx, repo.CreateChatLogParams{
		SessionID: sessionID,
		Action:    action,
		Detail:    sql.NullString{String: detail, Valid: detail != ""},
	})
	return err
}

func (m *Manager) CleanupOldKnowledge(ctx context.Context, olderThan time.Time) error {
	if _, err := m.repo.DeleteOldChatEntities(ctx, sql.NullTime{Time: olderThan, Valid: true}); err != nil {
		m.logger.Warnf("cleanup old chat entities error: %v", err)
	}
	if _, err := m.repo.DeleteOldChatRelations(ctx, sql.NullTime{Time: olderThan, Valid: true}); err != nil {
		m.logger.Warnf("cleanup old chat relations error: %v", err)
	}
	if _, err := m.repo.DeleteOldChatKnowledge(ctx, sql.NullTime{Time: olderThan, Valid: true}); err != nil {
		m.logger.Warnf("cleanup old chat knowledge error: %v", err)
	}
	return nil
}

var wikiStopWords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
	"is": true, "are": true, "was": true, "were": true, "be": true, "been": true,
	"have": true, "has": true, "had": true, "do": true, "does": true, "did": true,
	"will": true, "would": true, "could": true, "should": true, "may": true, "might": true,
	"can": true, "to": true, "of": true, "in": true, "for": true, "on": true,
	"with": true, "at": true, "by": true, "from": true, "as": true, "into": true,
	"through": true, "during": true, "before": true, "after": true, "above": true,
	"below": true, "between": true, "under": true, "again": true, "further": true,
	"then": true, "once": true, "here": true, "there": true, "when": true,
	"where": true, "why": true, "how": true, "all": true, "each": true, "few": true,
	"more": true, "most": true, "other": true, "some": true, "such": true,
	"no": true, "nor": true, "not": true, "only": true, "own": true, "same": true,
	"so": true, "than": true, "too": true, "very": true, "just": true, "also": true,
	"now": true, "this": true, "that": true, "these": true, "those": true,
	"i": true, "me": true, "my": true, "you": true, "your": true, "he": true,
	"she": true, "it": true, "we": true, "they": true, "what": true, "which": true,
	"who": true, "whom": true, "if": true, "because": true, "until": true,
	"while": true, "about": true, "against": true, "any": true, "both": true,
	"down": true, "up": true, "out": true, "off": true, "over": true,
	"hello": true, "hi": true, "hey": true, "thanks": true, "please": true, "sorry": true,
}

func isCJK(r rune) bool {
	return r >= 0x4E00 && r <= 0x9FFF
}

var englishWordRe = regexp.MustCompile(`[A-Za-z][A-Za-z0-9]*`)

func extractKeywordsFromQuery(query string) []string {
	if query == "" {
		return nil
	}

	isChineseQuery := false
	hasEnglish := false
	for _, r := range query {
		if isCJK(r) {
			isChineseQuery = true
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasEnglish = true
		}
	}

	var result []string
	seen := make(map[string]bool)

	if isChineseQuery && hasEnglish {
		for _, match := range englishWordRe.FindAllString(query, -1) {
			lower := strings.ToLower(match)
			if len(lower) >= 2 && !wikiStopWords[lower] && !seen[lower] {
				seen[lower] = true
				result = append(result, lower)
			}
		}
	}

	if len(result) == 0 {
		if isChineseQuery {
			trimmed := strings.TrimSpace(query)
			if len(trimmed) >= 2 {
				return []string{strings.ToLower(trimmed)}
			}
			return nil
		}
	}

	words := strings.Fields(query)
	for _, w := range words {
		w = strings.ToLower(w)
		w = strings.TrimSpace(w)
		if len(w) < 2 {
			continue
		}
		if wikiStopWords[w] {
			continue
		}
		if !seen[w] {
			seen[w] = true
			result = append(result, w)
		}
	}
	if len(result) > 6 {
		result = result[:6]
	}
	return result
}
