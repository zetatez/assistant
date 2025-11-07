
-- name: CreateTodoList :execresult
INSERT INTO todo_list (title, content, deadline)
VALUES (?, ?, ?);

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
  is_done
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
  is_done
FROM todo_list
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchTodoListsByTitle :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  title,
  content,
  deadline,
  is_done
FROM todo_list
WHERE title like ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchTodoListsByContent :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  title,
  content,
  deadline,
  is_done
FROM todo_list
WHERE content like ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchTodoListsByDeadlineLT :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  title,
  content,
  deadline,
  is_done
FROM todo_list
WHERE deadline < ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchTodoListsByTitleAndContent :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  title,
  content,
  deadline,
  is_done
FROM todo_list
WHERE title like ?
  AND content like ?
ORDER BY id
LIMIT ? OFFSET ?;

-- name: UpdateTodoListByID :execresult
UPDATE todo_list
SET  title = ?,
   content = ?,
  deadline = ?
WHERE id = ?
LIMIT 1;

-- name: MarkTodoListAsDoneByID :execresult
UPDATE todo_list
SET is_done = 1
WHERE id = ?
LIMIT 1;
