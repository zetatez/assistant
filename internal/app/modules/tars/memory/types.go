package memory

import "time"

type ShortTermMessage struct {
	ID      int64
	Role    string
	Content string
	Time    time.Time
}

type MemoryDoc struct {
	SessionID string
	Content   string
	Version   int
}

const (
	DefaultShortTermCapacity = 64

	DocTemplate = `# Memory Document

## User Profile
- **Name**: {username}
- **Last Updated**: {last_updated}

## Active Topics
(暂无活跃话题)

## Key Decisions & Conclusions
(暂无决策记录)

## Pending Tasks
(暂无待办事项)

## Important Context
(暂无重要上下文)

## Conversation Logs
(对话记录将自动追加于此)
`
)

type RecallResult struct {
	Content        string
	RelevanceScore float64
	SourceTime     time.Time
}

type UpdateReason struct {
	Type    string // "user_info" | "decision" | "task" | "context" | "topic"
	Content string
	Date    time.Time
}
