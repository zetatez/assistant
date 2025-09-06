package repo

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
)

type WikiRepo struct {
	db *Queries
}

func NewWikiRepo(db *Queries) *WikiRepo {
	return &WikiRepo{db: db}
}

func (r *WikiRepo) CreateEntry(ctx context.Context, title, content, keywords, createdBy string) (*Wiki, error) {
	hash := r.hashContent(content)
	if keywords == "" {
		keywords = extractKeywords(content)
	}

	result, err := r.db.CreateWiki(ctx, CreateWikiParams{
		Title:       title,
		Content:     content,
		Keywords:    keywords,
		ContentHash: hash,
		CreatedBy:   createdBy,
	})
	if err != nil {
		return nil, fmt.Errorf("create wiki entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	if id > 0 {
		entry, err := r.db.GetWikiByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("get created entry: %w", err)
		}
		return &entry, nil
	}

	existing, err := r.db.GetWikiByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("get existing entry: %w", err)
	}
	return &existing, nil
}

func (r *WikiRepo) GetByID(ctx context.Context, id int64) (*Wiki, error) {
	entry, err := r.db.GetWikiByID(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get entry by id: %w", err)
	}
	return &entry, nil
}

func (r *WikiRepo) List(ctx context.Context, offset, limit int) ([]*Wiki, error) {
	entries, err := r.db.ListWiki(ctx, ListWikiParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list wiki entries: %w", err)
	}

	result := make([]*Wiki, len(entries))
	for i := range entries {
		result[i] = &entries[i]
	}
	return result, nil
}

func (r *WikiRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.DeleteWikiByID(ctx, id)
	if err != nil {
		return fmt.Errorf("delete wiki entry: %w", err)
	}
	return nil
}

func (r *WikiRepo) Search(ctx context.Context, query string, limit int) ([]*Wiki, error) {
	if limit <= 0 {
		limit = 10
	}

	entries, err := r.db.SearchWiki(ctx, SearchWikiParams{
		Column1: query,
		Column2: query,
		Limit:   int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("search wiki entries: %w", err)
	}

	result := make([]*Wiki, len(entries))
	for i := range entries {
		result[i] = &entries[i]
	}
	return result, nil
}

func (r *WikiRepo) hashContent(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}

func extractKeywords(content string) string {
	words := strings.Fields(content)
	if len(words) < 3 {
		return strings.Join(words, ", ")
	}

	var keywords []string
	seen := make(map[string]bool)

	for _, word := range words {
		word = strings.ToLower(strings.Trim(word, ".,!?;:\"'()[]{}"))
		if len(word) > 2 && !seen[word] {
			keywords = append(keywords, word)
			seen[word] = true
			if len(keywords) >= 10 {
				break
			}
		}
	}
	return strings.Join(keywords, ", ")
}
