
-- name: AddTaskComposite :execresult
INSERT INTO task_composite (task_tmpl_id, task_atomic_id, step_order, depends_on)
VALUES (?, ?, ?, ?);

-- name: DeleteTaskComposite :execrows
DELETE FROM task_composite
WHERE task_tmpl_id = ?;

-- name: GetTaskComposite :many
SELECT id, gmt_create, gmt_modified, task_tmpl_id, task_atomic_id, step_order, depends_on
FROM task_composite
WHERE task_tmpl_id = ?
ORDER BY step_order ASC;

-- name: ListTaskComposite :many
SELECT id, gmt_create, gmt_modified, task_tmpl_id, task_atomic_id, step_order, depends_on
FROM task_composite
ORDER BY task_tmpl_id, step_order ASC
LIMIT ? OFFSET ?;

-- name: ListTaskAtomicByTaskTmplID :many
SELECT t1.task_tmpl_id, t1.step_order, t2.id, t2.gmt_create, t2.gmt_modified, t2.name, t2.description, t2.task_type, t2.config, t2.is_enable
FROM task_composite t1
JOIN task_atomic t2 ON t1.atomic_id = t2.id
WHERE t1.task_tmpl_id = ?
ORDER BY t1.step_order ASC;
