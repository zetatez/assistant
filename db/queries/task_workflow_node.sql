-- name: CreateTaskWorkflowNode :execresult
INSERT INTO task_workflow_node (
  workflow_id,
  node_id,
  node_type,
  display_name,
  task_atomic_def_id,
  sub_workflow_id,
  condition_expr,
  node_config,
  retry_policy,
  timeout,
  ord
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetTaskWorkflowNodeByID :one
SELECT * FROM task_workflow_node WHERE id = ?;

-- name: GetTaskWorkflowNodeByWorkflowAndNodeID :one
SELECT * FROM task_workflow_node WHERE workflow_id = ? AND node_id = ?;

-- name: ListTaskWorkflowNodes :many
SELECT * FROM task_workflow_node
WHERE workflow_id = ?
ORDER BY ord ASC;

-- name: DeleteTaskWorkflowNodeByWorkflowID :execresult
DELETE FROM task_workflow_node WHERE workflow_id = ?;
