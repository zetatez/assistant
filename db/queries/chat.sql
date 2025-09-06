-- name: CreateChatMessage :execresult
INSERT INTO chat_messages (chat_id, open_id, username, role, content, message_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, NOW());

-- name: GetChatMessages :many
SELECT id, chat_id, open_id, username, role, content, message_id, created_at
FROM chat_messages
WHERE chat_id = ?
ORDER BY created_at ASC
LIMIT ? OFFSET ?;

-- name: GetChatMessagesByTimeRange :many
SELECT id, chat_id, open_id, username, role, content, message_id, created_at
FROM chat_messages
WHERE chat_id = ? AND created_at >= ? AND created_at <= ?
ORDER BY created_at ASC;

-- name: CountChatMessages :one
SELECT count(*) FROM chat_messages WHERE chat_id = ?;

-- name: SearchChatMessagesByKeyword :many
SELECT id, chat_id, open_id, username, role, content, message_id, created_at
FROM chat_messages
WHERE chat_id = ? AND content LIKE ?
ORDER BY created_at DESC
LIMIT ?;

-- name: CreateChatMemory :execresult
INSERT INTO chat_memory (chat_id, keyword, summary, start_time, end_time, message_count, created_at)
VALUES (?, ?, ?, ?, ?, ?, NOW());

-- name: GetChatMemories :many
SELECT id, chat_id, keyword, summary, start_time, end_time, message_count, created_at
FROM chat_memory
WHERE chat_id = ?
ORDER BY created_at DESC
LIMIT ?;

-- name: SearchChatMemoriesByKeyword :many
SELECT id, chat_id, keyword, summary, start_time, end_time, message_count, created_at
FROM chat_memory
WHERE chat_id = ? AND keyword LIKE ?
ORDER BY created_at DESC
LIMIT ?;

-- name: DeleteOldChatMessages :execresult
DELETE FROM chat_messages WHERE created_at < DATE_SUB(NOW(), INTERVAL 90 DAY);

-- name: DeleteOldChatMemories :execresult
DELETE FROM chat_memory WHERE created_at < DATE_SUB(NOW(), INTERVAL 90 DAY);
