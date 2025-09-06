-- name: CreateTaskAtomicInstance :execresult
INSERT INTO task_atomic_instance (
  node_instance_id,
  task_atomic_def_id,
  status,
  input_params,
  timeout
) VALUES (?, ?, ?, ?, ?);

-- name: GetTaskAtomicInstanceByID :one
SELECT * FROM task_atomic_instance WHERE id = ?;

-- name: ListTaskAtomicInstances :many
SELECT * FROM task_atomic_instance
WHERE node_instance_id = ?
ORDER BY id ASC;

-- name: UpdateTaskAtomicInstanceStatus :execresult
UPDATE task_atomic_instance SET
  status = ?,
  output_result = ?,
  execution_log = ?,
  error_log = ?,
  gmt_start = COALESCE(?, gmt_start),
  gmt_end = COALESCE(?, gmt_end),
  duration_ms = ?,
  worker_id = ?
WHERE id = ?;

-- name: UpdateTaskAtomicInstanceRollback :execresult
UPDATE task_atomic_instance SET
  status = 'ROLLED_BACK',
  rollback_log = ?,
  rollback_result = ?
WHERE id = ?;
