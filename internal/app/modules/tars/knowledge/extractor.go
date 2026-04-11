package knowledge

import (
	"context"
	"encoding/json"
	"strings"

	"assistant/pkg/llm"
)

type Extractor struct {
	llmClient llm.Client
	llmModel  string
}

func NewExtractor(llmClient llm.Client, llmModel string) *Extractor {
	return &Extractor{
		llmClient: llmClient,
		llmModel:  llmModel,
	}
}

type ExtractedEntity struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
}

func (e *Extractor) ExtractEntities(ctx context.Context, messages []string) ([]Entity, error) {
	if e.llmClient == nil {
		return nil, nil
	}

	var sb strings.Builder
	for _, msg := range messages {
		sb.WriteString(msg)
		sb.WriteString("\n---\n")
	}

	prompt := `Extract key entities from the conversation below. Return as JSON array.

Entity types:
- person: person
- topic: topic/subject
- concept: concept/term
- event: event/activity

Format:
[{"type": "topic", "name": "distributed lock", "description": "a mechanism for mutual exclusion in distributed systems", "keywords": "lock,concurrency,mutex"}]

Return only JSON, no explanation.

Conversation:
` + sb.String()

	resp, err := e.llmClient.Chat(ctx, llm.ChatRequest{
		Model:       e.llmModel,
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: prompt}},
		Temperature: 0.3,
		MaxTokens:   500,
	})
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(resp.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var extracted []ExtractedEntity
	if err := json.Unmarshal([]byte(content), &extracted); err != nil {
		return nil, err
	}

	entities := make([]Entity, len(extracted))
	for i, ex := range extracted {
		entities[i] = Entity{
			Type:        EntityType(ex.Type),
			Name:        ex.Name,
			Description: ex.Description,
			Keywords:    ex.Keywords,
		}
	}

	return entities, nil
}

func (e *Extractor) ExtractRelations(ctx context.Context, entities []Entity, entityNameToID map[string]int64, newContent string) ([]Relation, error) {
	if e.llmClient == nil || len(entities) < 2 || len(entityNameToID) == 0 {
		return nil, nil
	}

	var sb strings.Builder
	for _, entity := range entities {
		sb.WriteString(string(entity.Type) + ": " + entity.Name + "\n")
	}

	prompt := `Given the following entities and new conversation content, identify relationships between entities. Return as JSON array.

Relationship types:
- related_to: related
- depends_on: dependency
- contradicts: contradiction
- part_of: containment

Format:
[{"from": "EntityA", "to": "EntityB", "type": "related_to", "context": "relationship context from conversation"}]

Return only JSON, no explanation.

Entity list:
` + sb.String() + "\n\nNew conversation:\n" + newContent

	resp, err := e.llmClient.Chat(ctx, llm.ChatRequest{
		Model:       e.llmModel,
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: prompt}},
		Temperature: 0.3,
		MaxTokens:   300,
	})
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(resp.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var extracted []struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Type    string `json:"type"`
		Context string `json:"context"`
	}
	if err := json.Unmarshal([]byte(content), &extracted); err != nil {
		return nil, err
	}

	relations := make([]Relation, 0, len(extracted))
	for _, ex := range extracted {
		fromID, ok := entityNameToID[ex.From]
		if !ok {
			continue
		}
		toID, ok := entityNameToID[ex.To]
		if !ok {
			continue
		}
		relations = append(relations, Relation{
			FromEntityID: fromID,
			ToEntityID:   toID,
			Type:         RelationType(ex.Type),
			Context:      ex.Context,
		})
	}

	return relations, nil
}
