-- name: CountTaskAtomicDefs :one
SELECT COUNT(*) AS cnt FROM task_atomic_def WHERE status = 'ENABLED';

-- name: GetTaskAtomicDefByID :one
SELECT * FROM task_atomic_def WHERE id = ?;

-- name: GetTaskAtomicDefByName :one
SELECT * FROM task_atomic_def WHERE name = ? LIMIT 1;

-- name: ListTaskAtomicDefs :many
SELECT * FROM task_atomic_def
WHERE status = 'ENABLED'
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: CreateTaskAtomicDef :execresult
INSERT INTO task_atomic_def (
  name,
  description,
  task_category,
  script_type,
  script_content,
  rollback_script_type,
  rollback_script_content,
  http_config,
  builtin_config,
  timeout,
  retry_count,
  retry_interval,
  is_rollback_supported,
  parameters,
  output_schema,
  env_vars,
  working_dir,
  status
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateTaskAtomicDefByID :execresult
UPDATE task_atomic_def SET
  name = ?,
  description = ?,
  task_category = ?,
  script_type = ?,
  script_content = ?,
  rollback_script_type = ?,
  rollback_script_content = ?,
  http_config = ?,
  builtin_config = ?,
  timeout = ?,
  retry_count = ?,
  retry_interval = ?,
  is_rollback_supported = ?,
  parameters = ?,
  output_schema = ?,
  env_vars = ?,
  working_dir = ?,
  status = ?
WHERE id = ?;

-- name: DeleteTaskAtomicDefByID :execresult
DELETE FROM task_atomic_def WHERE id = ? LIMIT 1;
