-- name: CreateTaskExecutionLog :execresult
INSERT INTO task_execution_log (
  workflow_instance_id,
  node_instance_id,
  task_atomic_instance_id,
  log_level,
  log_type,
  message,
  context
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListTaskExecutionLogs :many
SELECT * FROM task_execution_log
WHERE workflow_instance_id = ?
ORDER BY gmt_create ASC;
