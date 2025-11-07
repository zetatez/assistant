
-- name: CountTaskDefs :one
SELECT COUNT(*) AS cnt FROM task_def;

-- name: CreateTaskDef :execresult
INSERT INTO task_def (
  name,
  description,
  task_type,
  config,
  retry_policy
) VALUES (?, ?, ?, ?, ?);

-- name: DeleteTaskDefByID :execresult
DELETE FROM task_def WHERE id = ? LIMIT 1;

-- name: BatchDeleteTaskDefByID :execresult
DELETE FROM task_def WHERE id IN (sqlc.slice(ids));

-- name: GetTaskDefByID :one
SELECT
  id,
  gmt_create,
  gmt_modified,
  name,
  description,
  task_type,
  config,
  retry_policy
FROM task_def
WHERE id = ?;

-- name: GetTaskDefByName :one
SELECT
  id,
  gmt_create,
  gmt_modified,
  name,
  description,
  task_type,
  config,
  retry_policy
FROM task_def
WHERE name = ?
ORDER BY id DESC;

-- name: ListTaskDefs :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  name,
  description,
  task_type,
  config,
  retry_policy
FROM task_def
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: ListTaskDefsByType :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  name,
  description,
  task_type,
  config,
  retry_policy
FROM task_def
WHERE task_type = ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchTaskDefsByName :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  name,
  description,
  task_type,
  config,
  retry_policy
FROM task_def
WHERE name LIKE ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: UpdateTaskDefByID :execresult
UPDATE task_def
SET
  name = ?,
  description = ?,
  task_type = ?,
  config = ?,
  retry_policy = ?
WHERE id = ?
LIMIT 1;

-- name: UpdateTaskDefConfigByID :execresult
UPDATE task_def
SET config = ?, retry_policy = ?
WHERE id = ?;

