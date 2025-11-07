
-- name: CountTaskInstances :one
SELECT COUNT(*) AS cnt
FROM task_instance;

-- name: CountTaskInstancesByStatus :one
SELECT COUNT(*) AS cnt
FROM task_instance
WHERE status = ?;

-- name: CountTaskInstancesByDefID :one
SELECT COUNT(*) AS cnt
FROM task_instance
WHERE task_def_id = ?;

-- name: CreateTaskInstance :execresult
INSERT INTO task_instance (
  task_def_id,
  parent_instance_id,
  ord,
  status,
  result,
  err_msg
)
VALUES (?, ?, ?, ?, ?, ?);

-- name: DeleteTaskInstanceByID :execresult
DELETE FROM task_instance WHERE id = ? LIMIT 1;

-- name: DeleteTaskInstancesByParent :execresult
DELETE FROM task_instance WHERE parent_instance_id = ?;

-- name: GetTaskInstanceByID :one
SELECT
  id,
  gmt_create,
  gmt_modified,
  task_def_id,
  parent_instance_id,
  ord,
  status,
  result,
  err_msg
FROM task_instance
WHERE id = ?
LIMIT 1;

-- name: ListTaskInstances :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  task_def_id,
  parent_instance_id,
  ord,
  status,
  result,
  err_msg
FROM task_instance
ORDER BY ID DESC
LIMIT ? OFFSET ?;

-- name: UpdateTaskInstanceByID :execresult
UPDATE task_instance
SET status = ?,
    task_def_id = ?,
    parent_instance_id = ?,
    ord = ?,
    status = ?,
    result = ?,
    err_msg = ?
WHERE id = ?
LIMIT 1;

-- name: UpdateTaskInstanceStatusByID :execresult
UPDATE task_instance
SET status = ?
WHERE id = ?
LIMIT 1;

-- name: UpdateTaskInstanceResultByID :execresult
UPDATE task_instance
SET result = ?, err_msg = ?, status = ?
WHERE id = ?
LIMIT 1;

-- name: ResetTaskInstanceStatus :execresult
UPDATE task_instance
SET status = 'PENDING', result = NULL, err_msg = NULL
WHERE id = ?
LIMIT 1;

-- name: GetChildTaskInstances :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  task_def_id,
  parent_instance_id,
  ord,
  status,
  result,
  err_msg
FROM task_instance
WHERE parent_instance_id = ?
ORDER BY ord ASC;

-- name: GetRootTaskInstances :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  task_def_id,
  parent_instance_id,
  ord,
  status,
  result,
  err_msg
FROM task_instance
WHERE parent_instance_id IS NULL
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: GetTaskInstanceTree :many
WITH RECURSIVE instance_tree AS (
  SELECT
    ti.id,
    ti.task_def_id,
    ti.parent_instance_id,
    ti.ord,
    ti.status,
    ti.result,
    ti.err_msg,
    td.name AS task_name,
    td.task_type
  FROM task_instance ti
  JOIN task_def td ON ti.task_def_id = td.id
  WHERE ti.id = ?
  UNION ALL
  SELECT
    ti.id,
    ti.task_def_id,
    ti.parent_instance_id,
    ti.ord,
    ti.status,
    ti.result,
    ti.err_msg,
    td.name AS task_name,
    td.task_type
  FROM task_instance ti
  JOIN task_def td ON ti.task_def_id = td.id
  JOIN instance_tree it ON ti.parent_instance_id = it.id
)
SELECT * FROM instance_tree ORDER BY parent_instance_id, ord;

-- name: ListTaskInstancesByTaskDefID :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  task_def_id,
  parent_instance_id,
  ord,
  status,
  result,
  err_msg
FROM task_instance
WHERE task_def_id = ?
ORDER BY gmt_create DESC;

-- name: ListTaskInstancesByStatus :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  task_def_id,
  parent_instance_id,
  ord,
  status,
  result,
  err_msg
FROM task_instance
WHERE status = ?
ORDER BY gmt_create DESC
LIMIT ? OFFSET ?;

-- name: LockTaskInstanceForUpdate :one
SELECT
  id,
  gmt_create,
  gmt_modified,
  task_def_id,
  parent_instance_id,
  ord,
  status,
  result,
  err_msg
FROM task_instance
WHERE id = ?
FOR UPDATE;

-- name: LockPendingInstances :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  task_def_id,
  parent_instance_id,
  ord,
  status,
  result,
  err_msg
FROM task_instance
WHERE status = 'PENDING'
ORDER BY id DESC
LIMIT ?
FOR UPDATE SKIP LOCKED;

-- name: GetTaskInstanceSummary :one
SELECT
  SUM(CASE WHEN status = 'PENDING' THEN 1 ELSE 0 END) AS pending_count,
  SUM(CASE WHEN status = 'RUNNING' THEN 1 ELSE 0 END) AS running_count,
  SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) AS success_count,
  SUM(CASE WHEN status = 'FAILED' THEN 1 ELSE 0 END) AS failed_count,
  SUM(CASE WHEN status = 'CANCELLED' THEN 1 ELSE 0 END) AS cancelled_count
FROM task_instance;
