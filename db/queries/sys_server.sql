
-- name: CreateSysServer :execresult
INSERT IGNORE INTO sys_server (idc, svr_ip, svr_status, cpu_usage, mem_usage)
VALUES ('', ?, ?, 0, 0);

-- name: DeleteSysServerByID :execresult
DELETE FROM sys_server
WHERE id = ?
LIMIT 1;

-- name: DeleteSysServerBySvrIP :execresult
DELETE FROM sys_server
WHERE svr_ip = ?
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
  svr_status,
  cpu_usage,
  mem_usage
FROM sys_server
WHERE id = ?
LIMIT 1;

-- name: GetSysServerBySvrIP :one
SELECT
  id,
  gmt_create,
  gmt_modified,
  idc,
  svr_ip,
  svr_status,
  cpu_usage,
  mem_usage
FROM sys_server
WHERE svr_ip = ?
LIMIT 1;

-- name: ListSysServers :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  idc,
  svr_ip,
  svr_status,
  cpu_usage,
  mem_usage
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
  svr_status,
  cpu_usage,
  mem_usage
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
  svr_status,
  cpu_usage,
  mem_usage
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
  svr_status,
  cpu_usage,
  mem_usage
FROM sys_server
WHERE idc like ?
  AND svr_ip like ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: UpdateSysServerMetricsBySvrIP :execresult
UPDATE sys_server
SET cpu_usage = ?,
    mem_usage = ?,
    svr_status = ?
WHERE svr_ip = ?
LIMIT 1;
