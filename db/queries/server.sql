
-- name: CreateServer :execresult
INSERT INTO server (idc, svr_ip, ak, sk)
VALUES (?, ?, ?, ?);

-- name: DeleteServer :execresult
DELETE FROM server
WHERE id = ?
LIMIT 1;

-- name: CountServer :one
SELECT count(*) ct FROM server;

-- name: GetServerByID :one
SELECT id,
       gmt_create,
       gmt_modified,
       idc,
       svr_ip,
       ak,
       sk,
       svr_status
FROM server
WHERE id = ?
LIMIT 1;

-- name: ListServers :many
SELECT id,
       gmt_create,
       gmt_modified,
       idc,
       svr_ip,
       ak,
       sk,
       svr_status
FROM server
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchServersByIDC :many
SELECT id,
       gmt_create,
       gmt_modified,
       idc,
       svr_ip,
       ak,
       sk,
       svr_status
FROM server
WHERE idc = ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchServersBySvrIP :many
SELECT id,
       gmt_create,
       gmt_modified,
       idc,
       svr_ip,
       ak,
       sk,
       svr_status
FROM server
WHERE svr_ip like ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchServersByIDCAndSvrIP :many
SELECT id,
       gmt_create,
       gmt_modified,
       idc,
       svr_ip,
       ak,
       sk,
       svr_status
FROM server
WHERE idc like ?
  AND svr_ip like ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: UpdateServer :execresult
UPDATE server
SET  idc = ?,
  svr_ip = ?,
      ak = ?,
      sk = ?
WHERE id = ?
LIMIT 1;

-- name: UpdateServerSvrStatus :execresult
UPDATE server
SET svr_status = ?
WHERE id = ?
LIMIT 1;
