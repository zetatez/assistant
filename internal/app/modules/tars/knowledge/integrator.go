package knowledge

import (
	"context"
	"database/sql"
	"strings"
	"sync"

	"assistant/internal/app/repo"
	"assistant/pkg/llm"
)

type Integrator struct {
	repo      *repo.Queries
	db        *sql.DB
	llmClient llm.Client
	llmModel  string
	logger    Logger
}

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
}

func (i *Integrator) DB() *sql.DB {
	return i.db
}

func (i *Integrator) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return i.db.BeginTx(ctx, nil)
}

func (i *Integrator) WithTx(tx *sql.Tx) *repo.Queries {
	return i.repo.WithTx(tx)
}

func NewIntegrator(repo *repo.Queries, db *sql.DB, llmClient llm.Client, llmModel string, logger Logger) *Integrator {
	return &Integrator{
		repo:      repo,
		db:        db,
		llmClient: llmClient,
		llmModel:  llmModel,
		logger:    logger,
	}
}

func (i *Integrator) SaveEntity(ctx context.Context, entity *Entity) error {
	_, err := i.repo.CreateChatEntity(ctx, repo.CreateChatEntityParams{
		SessionID:   entity.SessionID,
		EntityType:  string(entity.Type),
		EntityName:  entity.Name,
		Description: sql.NullString{String: entity.Description, Valid: entity.Description != ""},
		Keywords:    sql.NullString{String: entity.Keywords, Valid: entity.Keywords != ""},
	})
	return err
}

func (i *Integrator) SaveEntitiesAndGetIDs(ctx context.Context, entities []Entity, tx *sql.Tx) (map[string]int64, error) {
	var ownedTx *sql.Tx
	if tx == nil {
		var err error
		ownedTx, err = i.db.BeginTx(ctx, nil)
		if err != nil {
			i.logger.Errorf("begin transaction error: %v", err)
			return nil, err
		}
		tx = ownedTx
	}
	txQueries := i.repo.WithTx(tx)

	nameToID := make(map[string]int64)
	for _, entity := range entities {
		if _, found := nameToID[entity.Name]; found {
			continue
		}

		existing, err := txQueries.SearchChatEntities(ctx, repo.SearchChatEntitiesParams{
			EntityName: entity.Name,
			Keywords:   sql.NullString{String: entity.Name, Valid: true},
			Limit:      10,
		})
		if err == nil && len(existing) > 0 {
			for _, e := range existing {
				if e.EntityName == entity.Name {
					nameToID[entity.Name] = e.ID
					break
				}
			}
		}

		if _, found := nameToID[entity.Name]; found {
			continue
		}

		result, err := txQueries.CreateChatEntity(ctx, repo.CreateChatEntityParams{
			SessionID:   entity.SessionID,
			EntityType:  string(entity.Type),
			EntityName:  entity.Name,
			Description: sql.NullString{String: entity.Description, Valid: entity.Description != ""},
			Keywords:    sql.NullString{String: entity.Keywords, Valid: entity.Keywords != ""},
		})
		if err != nil {
			if ownedTx != nil {
				ownedTx.Rollback()
			}
			i.logger.Errorf("save entity error: %v", err)
			return nil, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			if ownedTx != nil {
				ownedTx.Rollback()
			}
			i.logger.Errorf("get last insert id error: %v", err)
			return nil, err
		}
		nameToID[entity.Name] = id
	}

	if ownedTx != nil {
		if err := tx.Commit(); err != nil {
			i.logger.Errorf("commit transaction error: %v", err)
			return nil, err
		}
	}
	return nameToID, nil
}

func (i *Integrator) SaveEntitiesAndGetIDsTx(ctx context.Context, entities []Entity, q *repo.Queries) (map[string]int64, error) {
	nameToID := make(map[string]int64)
	for _, entity := range entities {
		if _, found := nameToID[entity.Name]; found {
			continue
		}

		existing, err := q.SearchChatEntities(ctx, repo.SearchChatEntitiesParams{
			EntityName: entity.Name,
			Keywords:   sql.NullString{String: entity.Name, Valid: true},
			Limit:      10,
		})
		if err == nil && len(existing) > 0 {
			for _, e := range existing {
				if e.EntityName == entity.Name {
					nameToID[entity.Name] = e.ID
					break
				}
			}
		}

		if _, found := nameToID[entity.Name]; found {
			continue
		}

		result, err := q.CreateChatEntity(ctx, repo.CreateChatEntityParams{
			SessionID:   entity.SessionID,
			EntityType:  string(entity.Type),
			EntityName:  entity.Name,
			Description: sql.NullString{String: entity.Description, Valid: entity.Description != ""},
			Keywords:    sql.NullString{String: entity.Keywords, Valid: entity.Keywords != ""},
		})
		if err != nil {
			i.logger.Errorf("save entity error: %v", err)
			return nil, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			i.logger.Errorf("get last insert id error: %v", err)
			return nil, err
		}
		nameToID[entity.Name] = id
	}
	return nameToID, nil
}

func (i *Integrator) SaveRelation(ctx context.Context, relation *Relation, tx *sql.Tx) error {
	q := i.repo
	if tx != nil {
		q = i.repo.WithTx(tx)
	}
	_, err := q.CreateChatRelation(ctx, repo.CreateChatRelationParams{
		SessionID:    relation.SessionID,
		FromEntityID: relation.FromEntityID,
		ToEntityID:   relation.ToEntityID,
		RelationType: string(relation.Type),
		Context:      sql.NullString{String: relation.Context, Valid: relation.Context != ""},
	})
	return err
}

func (i *Integrator) CreateKnowledgePage(ctx context.Context, page *KnowledgePage, tx *sql.Tx) error {
	q := i.repo
	if tx != nil {
		q = i.repo.WithTx(tx)
	}
	entityID := sql.NullInt64{Int64: page.EntityID, Valid: page.EntityID > 0}
	sourceMsgs := sql.NullString{String: page.SourceMessages, Valid: page.SourceMessages != ""}
	_, err := q.CreateChatKnowledge(ctx, repo.CreateChatKnowledgeParams{
		SessionID:      page.SessionID,
		EntityID:       entityID,
		Title:          page.Title,
		Content:        page.Content,
		SourceMessages: sourceMsgs,
		Version:        sql.NullInt32{Int32: 1, Valid: true},
		IsDraft:        sql.NullBool{Bool: false, Valid: true},
	})
	return err
}

func (i *Integrator) UpdateKnowledgePage(ctx context.Context, pageID int64, content, sourceMsgs string) error {
	_, err := i.repo.UpdateChatKnowledge(ctx, repo.UpdateChatKnowledgeParams{
		Content:        content,
		SourceMessages: sql.NullString{String: sourceMsgs, Valid: sourceMsgs != ""},
		ID:             pageID,
	})
	return err
}

func (i *Integrator) GenerateKnowledgeContent(ctx context.Context, entity *Entity, relatedMessages []string) (string, error) {
	if i.llmClient == nil {
		return "", nil
	}

	var sb strings.Builder
	for _, msg := range relatedMessages {
		sb.WriteString(msg)
		sb.WriteString("\n---\n")
	}

	prompt := `Based on the following conversation content, generate a structured knowledge page for entity "` + entity.Name + `" in Markdown format.

Requirements:
1. Use ## for title
2. Include entity definition and description
3. Summarize key points
4. List related knowledge points
5. Keep concise, 200-500 words
6. SECURITY: Remove all sensitive info - replace names with [USER], specific numbers/amounts with [REDACTED], passwords/tokens with [HIDDEN]
7. Only keep generic knowledge that is safe to share across different conversations

Conversation:
` + sb.String()

	resp, err := i.llmClient.Chat(ctx, llm.ChatRequest{
		Model:       i.llmModel,
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: prompt}},
		Temperature: 0.3,
		MaxTokens:   800,
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Content), nil
}

func (i *Integrator) SearchRelatedKnowledge(ctx context.Context, chatID string, keywords []string) ([]KnowledgePage, error) {
	if len(keywords) == 0 {
		return nil, nil
	}

	type pageResult struct {
		pages []KnowledgePage
	}
	results := make([]pageResult, len(keywords))
	var wg sync.WaitGroup

	for j, kw := range keywords {
		wg.Add(1)
		go func(idx int, pattern string) {
			defer wg.Done()
			pages, err := i.repo.SearchChatKnowledge(ctx, repo.SearchChatKnowledgeParams{
				Title:   pattern,
				Content: pattern,
				Limit:   5,
			})
			if err != nil {
				if i.logger != nil {
					i.logger.Warnf("search knowledge error: %v", err)
				}
				return
			}
			ps := make([]KnowledgePage, 0, len(pages))
			for _, p := range pages {
				entityID := int64(0)
				if p.EntityID.Valid {
					entityID = p.EntityID.Int64
				}
				ps = append(ps, KnowledgePage{
					ID:        p.ID,
					Title:     p.Title,
					Content:   p.Content,
					SessionID: p.SessionID,
					EntityID:  entityID,
				})
			}
			results[idx] = pageResult{pages: ps}
		}(j, "%"+kw+"%")
	}
	wg.Wait()

	seen := make(map[int64]bool)
	var allPages []KnowledgePage
	for _, res := range results {
		for _, p := range res.pages {
			if !seen[p.ID] {
				seen[p.ID] = true
				allPages = append(allPages, p)
				if p.EntityID > 0 && p.Content != "" {
					go func(eid int64, content string) {
						i.ReinforceEntityFromKnowledge(ctx, eid, content)
					}(p.EntityID, p.Content)
				}
			}
		}
	}

	return allPages, nil
}

func (i *Integrator) BuildWikiContext(pages []KnowledgePage) string {
	if len(pages) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Related Knowledge\n\n")

	for _, page := range pages {
		sb.WriteString("### " + page.Title + "\n\n")
		sb.WriteString(page.Content)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func (i *Integrator) ReinforceEntityFromKnowledge(ctx context.Context, entityID int64, knowledgeContent string) error {
	if knowledgeContent == "" || entityID == 0 {
		return nil
	}

	entity, err := i.repo.GetChatEntityByID(ctx, entityID)
	if err != nil {
		return err
	}

	descLen := len(knowledgeContent)
	if descLen > 500 {
		descLen = 500
	}
	newKeywords := entity.Keywords.String + " " + extractKeywordsFromContent(knowledgeContent)
	kwLen := len(newKeywords)
	if kwLen > 500 {
		kwLen = 500
	}
	_, err = i.repo.UpdateChatEntity(ctx, repo.UpdateChatEntityParams{
		EntityName:  entity.EntityName,
		Description: sql.NullString{String: knowledgeContent[:descLen], Valid: true},
		Keywords:    sql.NullString{String: strings.TrimSpace(newKeywords[:kwLen]), Valid: true},
		ID:          entityID,
	})
	return err
}

func extractKeywordsFromContent(content string) string {
	words := strings.Fields(content)
	if len(words) <= 10 {
		return strings.Join(words, ",")
	}
	return strings.Join(words[:10], ",")
}

func (i *Integrator) AppendLog(ctx context.Context, sessionID, action, detail string) error {
	_, err := i.repo.CreateChatLog(ctx, repo.CreateChatLogParams{
		SessionID: sessionID,
		Action:    action,
		Detail:    sql.NullString{String: detail, Valid: detail != ""},
	})
	return err
}
