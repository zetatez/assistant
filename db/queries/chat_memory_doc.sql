-- name: GetChatMemoryDoc :one
SELECT id, session_id, content, version, created_at, updated_at
FROM chat_memory_doc
WHERE session_id = ?;

-- name: UpsertChatMemoryDoc :execresult
INSERT INTO chat_memory_doc (session_id, content, version)
VALUES (?, ?, 1)
ON DUPLICATE KEY UPDATE
    content = VALUES(content),
    version = version + 1;

-- name: CreateChatRecall :execresult
INSERT INTO chat_recall (session_id, query, recalled_content, relevance_score)
VALUES (?, ?, ?, ?);

-- name: SearchChatRecall :many
SELECT id, session_id, query, recalled_content, relevance_score, created_at
FROM chat_recall
WHERE session_id = ? AND query LIKE ?
ORDER BY relevance_score DESC, created_at DESC
LIMIT ?;

-- name: SearchOldChatMessagesByKeyword :many
SELECT id, session_id, open_id, username, role, content, message_id, created_at
FROM chat_messages
WHERE session_id = ? AND content LIKE ? AND created_at < DATE_SUB(NOW(), INTERVAL 7 DAY)
ORDER BY created_at DESC
LIMIT ?;
