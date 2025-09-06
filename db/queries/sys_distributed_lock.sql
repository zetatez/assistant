
-- 尝试获取锁（原子操作，锁存在且未过期则获取失败）
-- 使用事务 + SELECT FOR UPDATE 确保并发安全
-- 流程: 先查询 -> 如锁存在且未过期则失败 -> 如锁不存在或已过期则插入/更新
-- name: TryAcquireLock :execresult
INSERT INTO sys_distributed_lock (lock_key, lock_holder, lock_ttl, expire_time)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    lock_holder = IF(is_active = 1 AND expire_time > NOW(), lock_holder, VALUES(lock_holder)),
    lock_ttl = IF(is_active = 1 AND expire_time > NOW(), lock_ttl, VALUES(lock_ttl)),
    expire_time = IF(is_active = 1 AND expire_time > NOW(), expire_time, VALUES(expire_time)),
    is_active = IF(is_active = 1 AND expire_time > NOW(), is_active, 1);

-- 释放锁（必须验证 lock_holder）
-- name: ReleaseLock :execresult
UPDATE sys_distributed_lock
SET is_active = 0
WHERE lock_key = ? AND lock_holder = ? AND is_active = 1;

-- 续期锁（必须验证 lock_holder）
-- name: RenewLock :execresult
UPDATE sys_distributed_lock
SET expire_time = ?
WHERE lock_key = ? AND lock_holder = ? AND is_active = 1 AND expire_time > NOW();

-- 检查锁是否被持有（未过期）
-- name: IsLockHeld :one
SELECT 1
FROM sys_distributed_lock
WHERE lock_key = ? AND is_active = 1 AND expire_time > NOW()
LIMIT 1;

-- 获取锁信息（仅活跃锁）
-- name: GetLockInfo :one
SELECT
    id,
    gmt_create,
    gmt_modified,
    lock_key,
    lock_holder,
    lock_ttl,
    expire_time,
    is_active
FROM sys_distributed_lock
WHERE lock_key = ? AND is_active = 1
LIMIT 1;

-- 获取锁信息（包含所有状态）
-- name: GetLockInfoAll :one
SELECT
    id,
    gmt_create,
    gmt_modified,
    lock_key,
    lock_holder,
    lock_ttl,
    expire_time,
    is_active
FROM sys_distributed_lock
WHERE lock_key = ?
LIMIT 1;

-- 强制释放锁（不需要 holder，用于管理员或异常清理）
-- name: ForceReleaseLock :execresult
UPDATE sys_distributed_lock
SET is_active = 0
WHERE lock_key = ?;

-- 统计活跃锁数量
-- name: CountActiveLocks :one
SELECT count(*) ct
FROM sys_distributed_lock
WHERE is_active = 1 AND expire_time > NOW();

-- 分页查询活跃锁
-- name: ListActiveLocks :many
SELECT
    id,
    gmt_create,
    gmt_modified,
    lock_key,
    lock_holder,
    lock_ttl,
    expire_time,
    is_active
FROM sys_distributed_lock
WHERE is_active = 1 AND expire_time > NOW()
ORDER BY expire_time ASC
LIMIT ? OFFSET ?;

-- 清理过期锁（保留5分钟缓冲期）
-- name: CleanExpiredLocks :execresult
DELETE FROM sys_distributed_lock
WHERE is_active = 0 OR expire_time < DATE_SUB(NOW(), INTERVAL 5 MINUTE);
