-- name: CreateChatKnowledge :execresult
INSERT INTO chat_knowledge (session_id, entity_id, title, content, source_messages, version, is_draft)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetChatKnowledge :many
SELECT id, session_id, entity_id, title, content, source_messages, version, is_draft, created_at, updated_at
FROM chat_knowledge
WHERE session_id = ?
ORDER BY updated_at DESC
LIMIT ? OFFSET ?;

-- name: GetChatKnowledgeByEntity :many
SELECT id, session_id, entity_id, title, content, source_messages, version, is_draft, created_at, updated_at
FROM chat_knowledge
WHERE entity_id = ?
ORDER BY version DESC;

-- name: SearchChatKnowledge :many
SELECT id, session_id, entity_id, title, content, source_messages, version, is_draft, created_at, updated_at
FROM chat_knowledge
WHERE (title LIKE ? OR content LIKE ?)
ORDER BY updated_at DESC
LIMIT ?;

-- name: UpdateChatKnowledge :execresult
UPDATE chat_knowledge
SET content = ?, source_messages = ?, version = version + 1, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteChatKnowledgeByChatID :execresult
DELETE FROM chat_knowledge WHERE session_id = ?;

-- name: DeleteOldChatKnowledge :execresult
DELETE FROM chat_knowledge WHERE created_at < ?;