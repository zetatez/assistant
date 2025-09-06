
-- name: CreateSysUser :execresult
INSERT INTO sys_user (user_name, password, email, is_internal)
VALUES (?, ?, ?, 0);

-- name: DeleteSysUserByID :execresult
DELETE FROM sys_user
WHERE id = ?
LIMIT 1;

-- name: CountSysUsers :one
SELECT count(*) ct FROM sys_user;

-- name: GetSysUserByID :one
SELECT
  id,
  gmt_create,
  gmt_modified,
  user_name,
  password,
  email,
  is_internal
FROM sys_user
WHERE id = ?
LIMIT 1;

-- name: ListSysUsers :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  user_name,
  password,
  email,
  is_internal
FROM sys_user
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchSysUsersByUserName :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  user_name,
  password,
  email,
  is_internal
FROM sys_user
WHERE user_name like ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchSysUsersByEmail :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  user_name,
  password,
  email,
  is_internal
FROM sys_user
WHERE email like ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: UpdateSysUserByID :execresult
UPDATE sys_user
SET user_name = ?,
     password = ?,
        email = ?
WHERE id = ?
LIMIT 1;
