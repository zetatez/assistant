
-- name: CreateSysServer :execresult
INSERT INTO sys_server (idc, svr_ip, ak, sk)
VALUES (?, ?, ?, ?);

-- name: DeleteSysServer :execresult
DELETE FROM sys_server
WHERE id = ?
LIMIT 1;

-- name: CountSysServers :one
SELECT count(*) ct FROM sys_server;

-- name: GetSysServerByID :one
SELECT
  id,
  gmt_create,
  gmt_modified,
  idc,
  svr_ip,
  ak,
  sk,
  svr_status
FROM sys_server
WHERE id = ?
LIMIT 1;

-- name: ListSysServers :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  idc,
  svr_ip,
  ak,
  sk,
  svr_status
FROM sys_server
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchSysServersByIDC :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  idc,
  svr_ip,
  ak,
  sk,
  svr_status
FROM sys_server
WHERE idc = ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchSysServersBySvrIP :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  idc,
  svr_ip,
  ak,
  sk,
  svr_status
FROM sys_server
WHERE svr_ip like ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchSysServersByIDCAndSvrIP :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  idc,
  svr_ip,
  ak,
  sk,
  svr_status
FROM sys_server
WHERE idc like ?
  AND svr_ip like ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: UpdateSysServer :execresult
UPDATE sys_server
SET  idc = ?,
  svr_ip = ?,
      ak = ?,
      sk = ?
WHERE id = ?
LIMIT 1;

-- name: UpdateSysServerSvrStatus :execresult
UPDATE sys_server
SET svr_status = ?
WHERE id = ?
LIMIT 1;
