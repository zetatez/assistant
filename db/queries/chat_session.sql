-- name: GetOrCreateChatSession :one
SELECT id, session_id, summary, pending_tasks, context, updated_at
FROM chat_session
WHERE session_id = ?;

-- name: UpsertChatSession :execresult
INSERT INTO chat_session (session_id, summary, pending_tasks, context)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    summary = VALUES(summary),
    pending_tasks = VALUES(pending_tasks),
    context = VALUES(context);