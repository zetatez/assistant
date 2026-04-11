-- name: CreateChatLog :execresult
INSERT INTO chat_log (session_id, action, detail)
VALUES (?, ?, ?);

-- name: GetChatLogs :many
SELECT id, session_id, action, detail, created_at
FROM chat_log
WHERE session_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetChatLogsByAction :many
SELECT id, session_id, action, detail, created_at
FROM chat_log
WHERE session_id = ? AND action = ?
ORDER BY created_at DESC
LIMIT ?;

-- name: DeleteChatLogsByChatID :execresult
DELETE FROM chat_log WHERE session_id = ?;