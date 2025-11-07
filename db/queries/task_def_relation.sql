
-- name: CountTaskDefRelations :one
SELECT COUNT(*) AS cnt
FROM task_def_relation;

-- name: CountTaskDefRelationsByParendID :one
SELECT COUNT(*) AS cnt
FROM task_def_relation
WHERE parent_id = ?;

-- name: CreateTaskDefRelation :execresult
INSERT INTO task_def_relation (
  parent_id,
  child_id,
  ord
) VALUES (?, ?, ?);

-- name: DeleteTaskDefRelationByID :execresult
DELETE FROM task_def_relation
WHERE id = ?;

-- name: DeleteTaskDefRelationByParentIDAndChildID :execresult
DELETE FROM task_def_relation
WHERE parent_id = ? AND child_id = ?;

-- name: DeleteTaskDefRelationsByParentID :execresult
DELETE FROM task_def_relation
WHERE parent_id = ?;

-- name: DeleteTaskDefRelationsByChildID :execresult
DELETE FROM task_def_relation
WHERE child_id = ?;

-- name: GetTaskDefRelationByID :one
SELECT
  id,
  gmt_create,
  gmt_modified,
  parent_id,
  child_id,
  ord
FROM task_def_relation
WHERE id = ?;

-- name: ListTaskDefRelations :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  parent_id,
  child_id,
  ord
FROM task_def_relation
ORDER BY ID
LIMIT ? OFFSET ?;

-- name: ListTaskDefRelationsByParent :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  parent_id,
  child_id,
  ord
FROM task_def_relation
WHERE parent_id = ?
ORDER BY ord ASC;

-- name: ListTaskDefRelationsByChild :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  parent_id,
  child_id,
  ord
FROM task_def_relation
WHERE child_id = ?
ORDER BY gmt_create DESC;

-- name: GetChildTaskDefs :many
SELECT
  td.id,
  td.name,
  td.description,
  td.task_type,
  td.config,
  td.retry_policy,
  tdr.ord
FROM task_def_relation AS tdr
JOIN task_def AS td ON td.id = tdr.child_id
WHERE tdr.parent_id = ?
ORDER BY tdr.ord ASC;

-- name: GetParentTaskDef :one
SELECT
  td.id,
  td.name,
  td.description,
  td.task_type,
  td.config,
  td.retry_policy
FROM task_def_relation AS tdr
JOIN task_def AS td ON td.id = tdr.parent_id
WHERE tdr.child_id = ?
LIMIT 1;

-- name: UpdateTaskDefRelationByID :execresult
UPDATE task_def_relation
SET parent_id = ?,
     child_id = ?,
          ord = ?
WHERE id = ?;

-- name: UpdateTaskDefRelationOrd :execresult
UPDATE task_def_relation
SET ord = ?
WHERE parent_id = ? AND child_id = ?;

-- name: UpdateTaskDefRelationParent :execresult
UPDATE task_def_relation
SET parent_id = ?
WHERE id = ?;

-- name: GetTaskDefTree :many
WITH RECURSIVE def_tree AS (
  SELECT
    td.id,
    td.name,
    td.description,
    td.task_type,
    td.config,
    td.retry_policy,
    tdr.parent_id,
    tdr.ord,
    0 AS depth
  FROM task_def_relation AS tdr
  JOIN task_def AS td ON td.id = tdr.child_id
  WHERE tdr.parent_id = ?
  UNION ALL
  SELECT
    ctd.id,
    ctd.name,
    ctd.description,
    ctd.task_type,
    ctd.config,
    ctd.retry_policy,
    cr.parent_id,
    cr.ord,
    dt.depth + 1
  FROM def_tree AS dt
  JOIN task_def_relation AS cr ON cr.parent_id = dt.id
  JOIN task_def AS ctd ON ctd.id = cr.child_id
)
SELECT * FROM def_tree ORDER BY depth, ord;

