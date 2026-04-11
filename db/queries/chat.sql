-- name: CreateChatMessage :execresult
INSERT INTO chat_messages (session_id, open_id, username, role, content, message_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, NOW());

-- name: GetChatMessages :many
SELECT id, session_id, open_id, username, role, content, message_id, created_at
FROM chat_messages
WHERE session_id = ?
ORDER BY created_at ASC
LIMIT ? OFFSET ?;

-- name: GetChatMessagesByTimeRange :many
SELECT id, session_id, open_id, username, role, content, message_id, created_at
FROM chat_messages
WHERE session_id = ? AND created_at >= ? AND created_at <= ?
ORDER BY created_at ASC;

-- name: GetChatMessagesBefore :many
SELECT id, session_id, open_id, username, role, content, message_id, created_at
FROM chat_messages
WHERE session_id = ? AND created_at <= ?
ORDER BY created_at DESC
LIMIT ?;

-- name: CountChatMessages :one
SELECT count(*) FROM chat_messages WHERE session_id = ?;

-- name: SearchChatMessagesByKeyword :many
SELECT id, session_id, open_id, username, role, content, message_id, created_at
FROM chat_messages
WHERE session_id = ? AND content LIKE ?
ORDER BY created_at DESC
LIMIT ?;

-- name: CreateChatMemory :execresult
INSERT INTO chat_memory (session_id, keyword, summary, memory_type, start_time, end_time, message_count, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, NOW());

-- name: GetChatMemories :many
SELECT id, session_id, keyword, summary, memory_type, start_time, end_time, message_count, created_at
FROM chat_memory
WHERE session_id = ? AND memory_type != 'session'
ORDER BY created_at DESC
LIMIT ?;

-- name: GetSessionLatestMemory :one
SELECT id, session_id, keyword, summary, memory_type, start_time, end_time, message_count, created_at
FROM chat_memory
WHERE session_id = ? AND memory_type = 'session'
ORDER BY created_at DESC
LIMIT 1;

-- name: SearchChatMemoriesByKeyword :many
SELECT id, session_id, keyword, summary, memory_type, start_time, end_time, message_count, created_at
FROM chat_memory
WHERE session_id = ? AND memory_type != 'session' AND keyword LIKE ?
ORDER BY created_at DESC
LIMIT ?;

-- name: UpdateChatMemoryType :execresult
UPDATE chat_memory SET memory_type = 'historical' WHERE session_id = ? AND memory_type = 'session';

-- name: DeleteOldChatMessages :execresult
DELETE FROM chat_messages WHERE created_at < DATE_SUB(NOW(), INTERVAL 90 DAY);

-- name: DeleteChatMessagesByChatID :execresult
DELETE FROM chat_messages WHERE session_id = ?;

-- name: DeleteOldChatMemories :execresult
DELETE FROM chat_memory WHERE created_at < DATE_SUB(NOW(), INTERVAL 90 DAY);

-- name: DeleteChatMemoriesByChatID :execresult
DELETE FROM chat_memory WHERE session_id = ?;

-- name: ListRecentChatIDs :many
SELECT DISTINCT session_id FROM chat_messages ORDER BY created_at DESC LIMIT ?;
