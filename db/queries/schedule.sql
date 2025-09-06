-- name: CountTaskSchedules :one
SELECT COUNT(*) AS cnt FROM task_schedule WHERE status != 'DISABLED';

-- name: GetTaskScheduleByID :one
SELECT * FROM task_schedule WHERE id = ?;

-- name: ListTaskSchedules :many
SELECT * FROM task_schedule
WHERE status != 'DISABLED'
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: ListDueTaskSchedules :many
SELECT * FROM task_schedule
WHERE status = 'ENABLED' AND next_execute_at <= NOW()
ORDER BY next_execute_at ASC
LIMIT ?;

-- name: CreateTaskSchedule :execresult
INSERT INTO task_schedule (
  name,
  description,
  workflow_def_id,
  schedule_type,
  cron_expr,
  interval_seconds,
  execute_at,
  input_params,
  status,
  created_by
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateTaskScheduleByID :execresult
UPDATE task_schedule SET
  name = ?,
  description = ?,
  schedule_type = ?,
  cron_expr = ?,
  interval_seconds = ?,
  execute_at = ?,
  input_params = ?,
  status = ?,
  next_execute_at = ?
WHERE id = ?;

-- name: UpdateTaskScheduleLastExecute :execresult
UPDATE task_schedule SET
  last_execute_id = ?,
  next_execute_at = ?
WHERE id = ?;

-- name: DeleteTaskScheduleByID :execresult
DELETE FROM task_schedule WHERE id = ? LIMIT 1;
