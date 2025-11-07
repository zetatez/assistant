
-- name: CreateUser :execresult
INSERT INTO user (user_name, password, email)
VALUES (?, ?, ?);

-- name: DeleteUserByID :execresult
DELETE FROM user
WHERE id = ?
LIMIT 1;

-- name: CountUsers :one
SELECT count(*) ct FROM user;

-- name: GetUserByID :one
SELECT
  id,
  gmt_create,
  gmt_modified,
  user_name,
  password,
  email
FROM user
WHERE id = ?
LIMIT 1;

-- name: ListUsers :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  user_name,
  password,
  email
FROM user
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchUsersByUserName :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  user_name,
  password,
  email
FROM user
WHERE user_name like ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchUsersByEmail :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  user_name,
  password,
  email
FROM user
WHERE email like ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: UpdateUserByID :execresult
UPDATE user
SET user_name = ?,
     password = ?,
        email = ?
WHERE id = ?
LIMIT 1;

