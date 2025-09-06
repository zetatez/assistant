-- name: CountTaskWorkflowInstances :one
SELECT COUNT(*) AS cnt FROM task_workflow_instance;

-- name: GetTaskWorkflowInstanceByID :one
SELECT * FROM task_workflow_instance WHERE id = ?;

-- name: ListTaskWorkflowInstances :many
SELECT * FROM task_workflow_instance
ORDER BY gmt_create DESC
LIMIT ? OFFSET ?;

-- name: ListPendingTaskWorkflowInstances :many
SELECT * FROM task_workflow_instance
WHERE status = 'PENDING'
ORDER BY priority DESC, gmt_create ASC
LIMIT ?;

-- name: CreateTaskWorkflowInstance :execresult
INSERT INTO task_workflow_instance (
  workflow_def_id,
  workflow_def_version,
  trigger_type,
  trigger_id,
  input_params,
  status,
  execution_mode,
  priority,
  created_by
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateTaskWorkflowInstanceStatus :execresult
UPDATE task_workflow_instance SET
  status = ?,
  status_reason = ?,
  current_node_id = ?,
  gmt_start = COALESCE(?, gmt_start),
  gmt_end = COALESCE(?, gmt_end),
  gmt_paused = COALESCE(?, gmt_paused),
  completed_nodes = ?,
  failed_nodes = ?,
  result_summary = ?,
  error_info = ?
WHERE id = ?;

-- name: UpdateTaskWorkflowInstanceProgress :execresult
UPDATE task_workflow_instance SET
  completed_nodes = ?,
  failed_nodes = ?,
  total_nodes = ?
WHERE id = ?;
