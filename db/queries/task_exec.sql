
-- name: CountTaskExecsByInstanceID :one
SELECT COUNT(*) AS cnt
FROM task_exec
WHERE task_instance_id = ?;

-- name: CreateTaskExec :execresult
INSERT INTO task_exec (
  task_instance_id,
  gmt_start,
  gmt_end,
  status,
  log
)
VALUES (?, ?, ?, ?, ?);

-- name: DeleteTaskExecByID :execresult
DELETE FROM task_exec WHERE id = ? LIMIT 1;

-- name: GetTaskExecByID :one
SELECT
  id,
  gmt_create,
  gmt_modified,
  task_instance_id,
  gmt_start,
  gmt_end,
  status,
  log
FROM task_exec
WHERE id = ?
LIMIT 1;

-- name: ListTaskExecs :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  task_instance_id,
  gmt_start,
  gmt_end,
  status,
  log
FROM task_exec
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: ListTaskExecsByTaskInstanceID :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  task_instance_id,
  gmt_start,
  gmt_end,
  status,
  log
FROM task_exec
WHERE task_instance_id = ?
ORDER BY ID DESC;

-- name: GetLastTaskExecByInstanceID :one
SELECT
  id,
  gmt_create,
  gmt_modified,
  task_instance_id,
  gmt_start,
  gmt_end,
  status,
  log
FROM task_exec
WHERE task_instance_id = ?
ORDER BY gmt_start DESC
LIMIT 1;

-- name: GetTaskExecLogByID :one
SELECT log FROM task_exec WHERE id = ?;

-- name: UpdateTaskExecByID :execresult
UPDATE task_exec
SET task_instance_id = ?,
           gmt_start = ?,
             gmt_end = ?,
              status = ?,
                 log = ?
WHERE id = ?;

-- name: UpdateTaskExecStatus :execresult
UPDATE task_exec
SET status = ?
WHERE id = ?;

-- name: MarkTaskExecSuccess :execresult
UPDATE task_exec
SET status = 'SUCCESS'
WHERE id = ?;

-- name: MarkTaskExecFailed :execresult
UPDATE task_exec
SET status = 'FAILED', log = CONCAT(log, '\n', ?)
WHERE id = ?;

-- name: AppendTaskExecLog :execresult
UPDATE task_exec
SET log = CONCAT(COALESCE(log, ''), '\n', ?)
WHERE id = ?;

-- name: CountTaskExecsStatus :one
SELECT
  SUM(CASE WHEN status = 'RUNNING' THEN 1 ELSE 0 END) AS running_count,
  SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) AS success_count,
  SUM(CASE WHEN status = 'FAILED' THEN 1 ELSE 0 END) AS failed_count
FROM task_exec;

-- name: CountTaskExecsStatusByTaskInstanceID :one
SELECT
  task_instance_id,
  SUM(CASE WHEN status = 'RUNNING' THEN 1 ELSE 0 END) AS running_count,
  SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) AS success_count,
  SUM(CASE WHEN status = 'FAILED' THEN 1 ELSE 0 END) AS failed_count
FROM task_exec
WHERE task_instance_id = ?
GROUP BY task_instance_id;

-- name: GetAvgTaskExecDurationByTaskInstanceID :one
SELECT
  task_instance_id,
  AVG(TIMESTAMPDIFF(SECOND, gmt_start, gmt_end)) AS avg_duration_sec
FROM task_exec
WHERE task_instance_id = ?
  AND gmt_start IS NOT NULL AND gmt_end IS NOT NULL
GROUP BY task_instance_id;

-- name: GetTaskExecDetailByTaskInstanceID :one
SELECT
  te.id AS exec_id,
  te.gmt_start,
  te.gmt_end,
  te.status AS exec_status,
  te.log,
  ti.id AS instance_id,
  ti.status AS instance_status,
  td.id AS def_id,
  td.name AS def_name,
  td.task_type
FROM task_exec te
JOIN task_instance ti ON te.task_instance_id = ti.id
JOIN task_def td ON ti.task_def_id = td.id
WHERE te.id = ?;

-- name: ListTaskExecDetailsByTaskDefID :many
SELECT
  te.id AS exec_id,
  te.gmt_start,
  te.gmt_end,
  te.status AS exec_status,
  ti.id AS instance_id,
  td.name AS def_name,
  td.task_type
FROM task_exec te
JOIN task_instance ti ON te.task_instance_id = ti.id
JOIN task_def td ON ti.task_def_id = td.id
WHERE td.id = ?
ORDER BY te.gmt_start DESC
LIMIT ? OFFSET ?;
