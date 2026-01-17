-- name: CreateTaskWorkflowEdge :execresult
INSERT INTO task_workflow_edge (
  workflow_id,
  from_node_id,
  to_node_id,
  edge_type,
  condition_expr
) VALUES (?, ?, ?, ?, ?);

-- name: ListTaskWorkflowEdges :many
SELECT * FROM task_workflow_edge
WHERE workflow_id = ?
ORDER BY id ASC;

-- name: ListTaskWorkflowEdgesByFromNode :many
SELECT * FROM task_workflow_edge
WHERE workflow_id = ? AND from_node_id = ?
ORDER BY id ASC;

-- name: DeleteTaskWorkflowEdgeByWorkflowID :execresult
DELETE FROM task_workflow_edge WHERE workflow_id = ?;
