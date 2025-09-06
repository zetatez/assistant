
-- name: CreateTaskAtomic :execresult
INSERT INTO task_atomic (name, description, task_type, config, is_enable)
VALUES (?, ?, ?, ?, ?);

-- name: DeleteTaskAtomic :execrows
DELETE FROM task_atomic
WHERE id = ?;

-- name: GetTaskAtomicByID :one
SELECT id, gmt_create, gmt_modified, name, description, task_type, config, is_enable
FROM task_atomic
WHERE id = ?;

-- name: GetTaskAtomicByName :one
SELECT id, gmt_create, gmt_modified, name, description, task_type, config, is_enable
FROM task_atomic
WHERE name = ?;

-- name: ListTaskAtomics :many
SELECT id, gmt_create, gmt_modified, name, description, task_type, config, is_enable
FROM task_atomic
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchTaskAtomics :many
SELECT id, gmt_create, gmt_modified, name, description, task_type, config, is_enable
FROM task_atomic
WHERE name like ?
  AND description like ?
  AND task_type = ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: UpdateTaskAtomic :execrows
UPDATE task_atomic
SET
  name = ?,
  description = ?,
  task_type = ?,
  config = ?,
  is_enable = ?
WHERE id = ?;

