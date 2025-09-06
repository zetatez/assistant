-- name: CreateTaskNodeInstance :execresult
INSERT INTO task_node_instance (
  workflow_instance_id,
  node_def_id,
  node_id,
  status,
  input_params
) VALUES (?, ?, ?, ?, ?);

-- name: GetTaskNodeInstanceByID :one
SELECT * FROM task_node_instance WHERE id = ?;

-- name: ListTaskNodeInstances :many
SELECT * FROM task_node_instance
WHERE workflow_instance_id = ?
ORDER BY id ASC;

-- name: ListPendingTaskNodeInstances :many
SELECT * FROM task_node_instance
WHERE workflow_instance_id = ? AND status = 'PENDING'
ORDER BY id ASC;

-- name: UpdateTaskNodeInstanceStatus :execresult
UPDATE task_node_instance SET
  status = ?,
  status_reason = ?,
  output_result = ?,
  execution_log = ?,
  error_log = ?,
  gmt_start = COALESCE(?, gmt_start),
  gmt_end = COALESCE(?, gmt_end),
  duration_ms = ?,
  worker_id = ?
WHERE id = ?;

-- name: UpdateTaskNodeInstanceRetry :execresult
UPDATE task_node_instance SET
  retry_count = ?,
  status = 'PENDING'
WHERE id = ?;
