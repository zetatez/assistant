-- name: CreateChatEntity :execresult
INSERT INTO chat_entities (session_id, entity_type, entity_name, description, keywords)
VALUES (?, ?, ?, ?, ?);

-- name: GetChatEntities :many
SELECT id, session_id, entity_type, entity_name, description, keywords, created_at, updated_at
FROM chat_entities
WHERE session_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetChatEntitiesByType :many
SELECT id, session_id, entity_type, entity_name, description, keywords, created_at, updated_at
FROM chat_entities
WHERE session_id = ? AND entity_type = ?
ORDER BY entity_name;

-- name: SearchChatEntities :many
SELECT id, session_id, entity_type, entity_name, description, keywords, created_at, updated_at
FROM chat_entities
WHERE entity_name LIKE ? OR keywords LIKE ?
ORDER BY created_at DESC
LIMIT ?;

-- name: UpdateChatEntity :execresult
UPDATE chat_entities
SET entity_name = ?, description = ?, keywords = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: GetChatEntityByID :one
SELECT id, session_id, entity_type, entity_name, description, keywords, created_at, updated_at
FROM chat_entities
WHERE id = ?;

-- name: DeleteChatEntitiesByChatID :execresult
DELETE FROM chat_entities WHERE session_id = ?;

-- name: DeleteOldChatEntities :execresult
DELETE FROM chat_entities WHERE created_at < ?;