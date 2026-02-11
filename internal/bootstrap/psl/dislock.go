package psl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrLockNotFoundOrNotOwned = errors.New("lock not found or not owned by this instance")

var (
	disLocker     *DisLocker
	onceDisLocker sync.Once
)

func GetDisLocker() *DisLocker {
	if err := InitDisLocker(context.Background()); err != nil {
		GetLogger().Fatalf("init distributed locker failed: %v", err)
	}
	return disLocker
}

type DisLocker struct {
	db *sql.DB
}

type DisLockInfo struct {
	ID         int64
	LockKey    string
	LockValue  string
	ExpireTime time.Time
	CreateTime time.Time
	UpdateTime time.Time
}

func InitDisLocker(ctx context.Context) error {
	var initErr error
	onceDisLocker.Do(func() {
		if err := createSysDistributedLockTable(ctx, GetDB()); err != nil {
			initErr = err
			return
		}
		disLocker = &DisLocker{
			db: GetDB(),
		}
	})
	if initErr != nil {
		return initErr
	}
	return nil
}

func createSysDistributedLockTable(ctx context.Context, db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS sys_distributed_lock (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			lock_key VARCHAR(255) NOT NULL COMMENT '锁的唯一标识',
			lock_value VARCHAR(255) NOT NULL COMMENT '锁的值，用于标识锁的持有者',
			expire_time TIMESTAMP NOT NULL COMMENT '锁的过期时间',
			PRIMARY KEY (id),
			UNIQUE KEY uk_lock_key (lock_key),
			KEY idx_expire_time (expire_time)
		) COMMENT='分布式锁表';
	`

	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create sys_distributed_lock table: %w", err)
	}
	return nil
}

func normalizeTTLSeconds(ttl int) int {
	if ttl <= 0 {
		cfg := GetConfig()
		ttl = cfg.Dislock.DefaultTTL
		if ttl <= 0 {
			ttl = 30
		}
	}

	cfg := GetConfig()
	if cfg.Dislock.MaxTTL > 0 && ttl > cfg.Dislock.MaxTTL {
		ttl = cfg.Dislock.MaxTTL
	}
	return ttl
}

func (l *DisLocker) Lock(key, value string, ttl int) (bool, error) {
	ttl = normalizeTTLSeconds(ttl)

	query := `
		INSERT INTO sys_distributed_lock (lock_key, lock_value, expire_time)
		VALUES (?, ?, DATE_ADD(NOW(), INTERVAL ? SECOND))
		ON DUPLICATE KEY UPDATE
			lock_value = IF(expire_time < NOW(), VALUES(lock_value), lock_value),
			expire_time = IF(expire_time < NOW(), VALUES(expire_time), expire_time)
	`

	result, err := l.db.ExecContext(context.Background(), query, key, value, ttl)
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to check lock result: %w", err)
	}

	// If rows affected > 0, we either inserted or updated an expired lock
	return rowsAffected > 0, nil
}

func (l *DisLocker) Unlock(key, value string) error {
	query := `
		DELETE FROM sys_distributed_lock
		WHERE lock_key = ? AND lock_value = ?
	`

	result, err := l.db.ExecContext(context.Background(), query, key, value)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check unlock result: %w", err)
	}

	if rowsAffected == 0 {
		return ErrLockNotFoundOrNotOwned
	}

	return nil
}

func (l *DisLocker) Renew(key, value string, ttl int) (bool, error) {
	ttl = normalizeTTLSeconds(ttl)

	query := `
		UPDATE sys_distributed_lock
		SET expire_time = DATE_ADD(NOW(), INTERVAL ? SECOND)
		WHERE lock_key = ? AND lock_value = ? AND expire_time > NOW()
	`

	result, err := l.db.ExecContext(context.Background(), query, ttl, key, value)
	if err != nil {
		return false, fmt.Errorf("failed to renew lock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to check renew result: %w", err)
	}

	return rowsAffected > 0, nil
}

func (l *DisLocker) IsLocked(key string) (bool, error) {
	query := `
		SELECT 1 FROM sys_distributed_lock
		WHERE lock_key = ? AND expire_time > NOW()
		LIMIT 1
	`

	var one int
	err := l.db.QueryRowContext(context.Background(), query, key).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check lock status: %w", err)
	}
	return true, nil
}

func (l *DisLocker) GetDisLockInfo(key string) (*DisLockInfo, error) {
	query := `
		SELECT id, lock_key, lock_value, expire_time, gmt_create, gmt_modified
		FROM sys_distributed_lock
		WHERE lock_key = ? AND expire_time > NOW()
		LIMIT 1
	`

	var lock DisLockInfo
	err := l.db.QueryRowContext(context.Background(), query, key).Scan(
		&lock.ID,
		&lock.LockKey,
		&lock.LockValue,
		&lock.ExpireTime,
		&lock.CreateTime,
		&lock.UpdateTime,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get lock info: %w", err)
	}

	return &lock, nil
}

func (l *DisLocker) ForceUnlock(key string) error {
	query := `DELETE FROM sys_distributed_lock WHERE lock_key = ?`

	_, err := l.db.ExecContext(context.Background(), query, key)
	if err != nil {
		return fmt.Errorf("failed to force unlock: %w", err)
	}

	return nil
}

func (l *DisLocker) CleanExpiredLocks() (int64, error) {
	query := `DELETE FROM sys_distributed_lock WHERE expire_time <= NOW()`

	result, err := l.db.ExecContext(context.Background(), query)
	if err != nil {
		return 0, fmt.Errorf("failed to clean expired locks: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return rowsAffected, nil
}

// StartExpiredLockCleaner starts a background goroutine that periodically deletes expired locks.
// Call the returned cancel function to stop it.
func (l *DisLocker) StartExpiredLockCleaner(parent context.Context, interval time.Duration) context.CancelFunc {
	if interval <= 0 {
		interval = time.Minute
	}

	ctx, cancel := context.WithCancel(parent)
	go func() {
		// best-effort initial cleanup to keep the table small.
		if n, err := l.CleanExpiredLocks(); err != nil {
			GetLogger().Warnf("failed to clean expired locks: %v", err)
		} else if n > 0 {
			GetLogger().Infof("cleaned %d expired locks", n)
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if n, err := l.CleanExpiredLocks(); err != nil {
					GetLogger().Warnf("failed to clean expired locks: %v", err)
				} else if n > 0 {
					GetLogger().Infof("cleaned %d expired locks", n)
				}
			}
		}
	}()

	return cancel
}
