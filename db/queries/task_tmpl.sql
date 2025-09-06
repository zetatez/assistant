
-- name: CreateCompositeTask :execresult
INSERT INTO task_tmpl (name, description, config, is_enable)
VALUES (?, ?, ?, ?);

-- name: DeleteCompositeTask :execrows
DELETE FROM task_tmpl
WHERE id = ?;

-- name: GetCompositeTaskByID :one
SELECT id, gmt_create, gmt_modified, name, description, config, is_enable
FROM task_tmpl
WHERE id = ?;

-- name: GetCompositeTaskByName :one
SELECT id, gmt_create, gmt_modified, name, description, config, is_enable
FROM task_tmpl
WHERE name = ?;

-- name: ListCompositeTasks :many
SELECT id, gmt_create, gmt_modified, name, description, config, is_enable
FROM task_tmpl
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: UpdateCompositeTask :execrows
UPDATE task_tmpl
SET
  name = ?,
  description = ?,
  config = ?,
  is_enable = ?
WHERE id = ?;

