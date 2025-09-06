-- name: CountTaskWorkflowDefs :one
SELECT COUNT(*) AS cnt FROM task_workflow_def WHERE status != 'DEPRECATED';

-- name: GetTaskWorkflowDefByID :one
SELECT * FROM task_workflow_def WHERE id = ?;

-- name: GetTaskWorkflowDefByName :one
SELECT * FROM task_workflow_def WHERE name = ? ORDER BY version DESC LIMIT 1;

-- name: ListTaskWorkflowDefs :many
SELECT * FROM task_workflow_def
WHERE status != 'DEPRECATED'
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: CreateTaskWorkflowDef :execresult
INSERT INTO task_workflow_def (
  name,
  description,
  version,
  workflow_type,
  graph_config,
  parameters,
  timeout,
  on_error_strategy,
  notification_config,
  status,
  created_by
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateTaskWorkflowDefByID :execresult
UPDATE task_workflow_def SET
  name = ?,
  description = ?,
  workflow_type = ?,
  graph_config = ?,
  parameters = ?,
  timeout = ?,
  on_error_strategy = ?,
  notification_config = ?,
  status = ?
WHERE id = ?;

-- name: DeleteTaskWorkflowDefByID :execresult
DELETE FROM task_workflow_def WHERE id = ? LIMIT 1;
