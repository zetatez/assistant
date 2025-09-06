
-- name: CreateTodoList :execresult
INSERT INTO todo_list (title, content, deadline, progress, priority, task_status)
VALUES (?, ?, ?, 0, 5, 'PENDING');

-- name: DeleteTodoListByID :execresult
DELETE FROM todo_list
WHERE id = ?
LIMIT 1;

-- name: CountTodoList :one
SELECT count(*) ct FROM todo_list;

-- name: GetTodoListByID :one
SELECT
  id,
  gmt_create,
  gmt_modified,
  title,
  content,
  deadline,
  progress,
  priority,
  task_status
FROM todo_list
WHERE id = ?
LIMIT 1;

-- name: ListTodoLists :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  title,
  content,
  deadline,
  progress,
  priority,
  task_status
FROM todo_list
ORDER BY priority DESC, id DESC
LIMIT ? OFFSET ?;

-- name: SearchTodoListsByTitle :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  title,
  content,
  deadline,
  progress,
  priority,
  task_status
FROM todo_list
WHERE title like ?
ORDER BY priority DESC, id DESC
LIMIT ? OFFSET ?;

-- name: SearchTodoListsByContent :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  title,
  content,
  deadline,
  progress,
  priority,
  task_status
FROM todo_list
WHERE content like ?
ORDER BY priority DESC, id DESC
LIMIT ? OFFSET ?;

-- name: SearchTodoListsByDeadlineLT :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  title,
  content,
  deadline,
  progress,
  priority,
  task_status
FROM todo_list
WHERE deadline < ?
ORDER BY priority DESC, id DESC
LIMIT ? OFFSET ?;

-- name: SearchTodoListsByTitleAndContent :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  title,
  content,
  deadline,
  progress,
  priority,
  task_status
FROM todo_list
WHERE title like ?
  AND content like ?
ORDER BY priority, id
LIMIT ? OFFSET ?;

-- name: UpdateTodoListByID :execresult
UPDATE todo_list
SET  title = ?,
   content = ?,
   deadline = ?,
   progress = ?,
   priority = ?,
   task_status = ?
WHERE id = ?
LIMIT 1;

-- name: UpdateTodoListProgressByID :execresult
UPDATE todo_list
SET progress = ?,
    task_status = CASE WHEN ? >= 100 THEN 'COMPLETED' ELSE 'IN_PROGRESS' END
WHERE id = ?
LIMIT 1;

-- name: CompleteTodoListByID :execresult
UPDATE todo_list
SET progress = 100,
    task_status = 'COMPLETED'
WHERE id = ?
LIMIT 1;

-- name: UpdateTodoListPriorityByID :execresult
UPDATE todo_list
SET priority = ?
WHERE id = ?
LIMIT 1;
