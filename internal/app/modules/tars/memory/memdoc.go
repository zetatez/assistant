package memory

import (
	"context"
	"strings"
	"time"

	"assistant/internal/app/repo"
)

type Memdoc struct {
	repo   *repo.Queries
	logger Logger
}

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
}

func NewMemdoc(repo *repo.Queries, logger Logger) *Memdoc {
	return &Memdoc{
		repo:   repo,
		logger: logger,
	}
}

func (m *Memdoc) GetDoc(ctx context.Context, sessionID string) (*MemoryDoc, error) {
	doc, err := m.repo.GetChatMemoryDoc(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return &MemoryDoc{
		SessionID: doc.SessionID,
		Content:   doc.Content,
		Version:   int(doc.Version),
	}, nil
}

func (m *Memdoc) SaveDoc(ctx context.Context, sessionID, content string) error {
	_, err := m.repo.UpsertChatMemoryDoc(ctx, repo.UpsertChatMemoryDocParams{
		SessionID: sessionID,
		Content:   content,
	})
	return err
}

func (m *Memdoc) GetOrCreateDoc(ctx context.Context, sessionID, username string) (*MemoryDoc, error) {
	doc, err := m.GetDoc(ctx, sessionID)
	if err == nil {
		return doc, nil
	}

	if !strings.Contains(err.Error(), "no rows") {
		return nil, err
	}

	newDoc := m.createNewDoc(username)
	if err := m.SaveDoc(ctx, sessionID, newDoc.Content); err != nil {
		return nil, err
	}

	return newDoc, nil
}

func (m *Memdoc) createNewDoc(username string) *MemoryDoc {
	content := DocTemplate
	content = strings.ReplaceAll(content, "{username}", username)
	content = strings.ReplaceAll(content, "{last_updated}", time.Now().Format("2006-01-02 15:04"))
	return &MemoryDoc{
		Content: content,
	}
}

func (m *Memdoc) UpdateUserProfile(ctx context.Context, sessionID, username string) error {
	doc, err := m.GetOrCreateDoc(ctx, sessionID, username)
	if err != nil {
		return err
	}

	lines := strings.Split(doc.Content, "\n")
	var newLines []string
	inProfile := false
	for _, line := range lines {
		if strings.Contains(line, "## User Profile") {
			inProfile = true
			newLines = append(newLines, line)
			continue
		}
		if inProfile && strings.HasPrefix(strings.TrimSpace(line), "## ") {
			inProfile = false
		}
		if inProfile && strings.Contains(line, "**Last Updated**") {
			continue
		}
		if !inProfile {
			newLines = append(newLines, line)
		}
	}

	var sb strings.Builder
	for i, line := range newLines {
		if i == 1 {
			sb.WriteString("- **Last Updated**: " + time.Now().Format("2006-01-02 15:04") + "\n")
		}
		sb.WriteString(line)
		if i < len(newLines)-1 {
			sb.WriteString("\n")
		}
	}

	doc.Content = sb.String()
	return m.SaveDoc(ctx, sessionID, doc.Content)
}

func (m *Memdoc) AppendConversationLog(ctx context.Context, sessionID, topic, summary string) error {
	doc, err := m.GetOrCreateDoc(ctx, sessionID, "")
	if err != nil {
		return err
	}

	logEntry := strings.Builder{}
	logEntry.WriteString("\n### ")
	logEntry.WriteString(time.Now().Format("2006-01-02"))
	logEntry.WriteString(" - ")
	logEntry.WriteString(topic)
	logEntry.WriteString("\n")
	logEntry.WriteString(summary)
	logEntry.WriteString("\n")

	lines := strings.Split(doc.Content, "## Conversation Logs")
	if len(lines) == 2 {
		doc.Content = lines[0] + "## Conversation Logs" + logEntry.String() + lines[1]
	} else {
		doc.Content += logEntry.String()
	}

	return m.SaveDoc(ctx, sessionID, doc.Content)
}

func (m *Memdoc) SearchInDoc(doc *MemoryDoc, keyword string) bool {
	if doc == nil {
		return false
	}
	return strings.Contains(strings.ToLower(doc.Content), strings.ToLower(keyword))
}
