package psl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"assistant/internal/app/repo"
)

var (
	ErrLockNotFound    = errors.New("lock not found")
	ErrLockNotOwned    = errors.New("lock not owned by this holder")
	ErrLockExpired     = errors.New("lock has expired")
	ErrLockAlreadyHeld = errors.New("lock already held by another holder")
	ErrLockNotActive   = errors.New("lock is not active")
)

var (
	distributedLock     *DistributedLock
	onceDistributedLock sync.Once
)

func GetDistributedLock() *DistributedLock {
	if distributedLock == nil {
		GetLogger().Errorf("[distributed_lock] not initialized, returning nil")
	}
	return distributedLock
}

type DistributedLock struct {
	db         *sql.DB
	q          *repo.Queries
	cfg        *Config
	defaultTTL int
	maxTTL     int
}

type LockInfo struct {
	ID          int64     `json:"id"`
	GmtCreate   time.Time `json:"gmt_create"`
	GmtModified time.Time `json:"gmt_modified"`
	LockKey     string    `json:"lock_key"`
	LockHolder  string    `json:"lock_holder"`
	LockTTL     int       `json:"lock_ttl"`
	ExpireTime  time.Time `json:"expire_time"`
	IsActive    bool      `json:"is_active"`
}

func InitDistributedLock(ctx context.Context) error {
	var initErr error
	onceDistributedLock.Do(func() {
		if GetDB() == nil {
			initErr = fmt.Errorf("database not initialized")
			return
		}
		if err := createSysDistributedLockTable(ctx, GetDB()); err != nil {
			initErr = err
			return
		}
		cfg := GetConfig()
		distributedLock = &DistributedLock{
			db:         GetDB(),
			q:          repo.New(GetDB()),
			cfg:        cfg,
			defaultTTL: cfg.Dislock.DefaultTTL,
			maxTTL:     cfg.Dislock.MaxTTL,
		}
		if distributedLock.defaultTTL <= 0 {
			distributedLock.defaultTTL = 30
		}
	})
	return initErr
}

func createSysDistributedLockTable(ctx context.Context, db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS sys_distributed_locker (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			lock_key VARCHAR(255) NOT NULL DEFAULT '' COMMENT '锁的唯一标识',
			lock_holder VARCHAR(255) NOT NULL DEFAULT '' COMMENT '锁持有者标识',
			lock_ttl INT NOT NULL DEFAULT 0 COMMENT '锁的存活时间（秒）',
			expire_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '锁过期时间',
			is_active TINYINT NOT NULL DEFAULT 1 COMMENT '锁是否激活：1=激活，0=已释放',
			PRIMARY KEY (id),
			UNIQUE KEY uk_lock_key (lock_key),
			KEY idx_expire_time (expire_time),
			KEY idx_lock_holder (lock_holder)
		) COMMENT='分布式锁表';
	`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create sys_distributed_locker table: %w", err)
	}
	return nil
}

func (l *DistributedLock) normalizeTTL(ttl int) int {
	if ttl <= 0 {
		ttl = l.defaultTTL
	}
	if l.maxTTL > 0 && ttl > l.maxTTL {
		ttl = l.maxTTL
	}
	return ttl
}

func (l *DistributedLock) TryAcquire(ctx context.Context, key, holder string, ttl int) (acquired bool, err error) {
	ttl = l.normalizeTTL(ttl)
	expireTime := time.Now().Add(time.Duration(ttl) * time.Second)

	result, err := l.q.TryAcquireLock(ctx, repo.TryAcquireLockParams{
		LockKey:    key,
		LockHolder: holder,
		LockTtl:    int32(ttl),
		ExpireTime: expireTime,
	})
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to check acquire result: %w", err)
	}
	return rowsAffected > 0, nil
}

func (l *DistributedLock) Release(ctx context.Context, key, holder string) (released bool, err error) {
	result, err := l.q.ReleaseLock(ctx, repo.ReleaseLockParams{
		LockKey:    key,
		LockHolder: holder,
	})
	if err != nil {
		return false, fmt.Errorf("failed to release lock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to check release result: %w", err)
	}

	if rowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func (l *DistributedLock) Renew(ctx context.Context, key, holder string, ttl int) (renewed bool, err error) {
	ttl = l.normalizeTTL(ttl)
	expireTime := time.Now().Add(time.Duration(ttl) * time.Second)

	result, err := l.q.RenewLock(ctx, repo.RenewLockParams{
		ExpireTime: expireTime,
		LockKey:    key,
		LockHolder: holder,
	})
	if err != nil {
		return false, fmt.Errorf("failed to renew lock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to check renew result: %w", err)
	}
	return rowsAffected > 0, nil
}

func (l *DistributedLock) IsHeld(ctx context.Context, key string) (held bool, holder string, err error) {
	lock, err := l.q.GetLockInfo(ctx, key)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to get lock info: %w", err)
	}
	if lock.LockKey == "" {
		return false, "", nil
	}
	if lock.IsActive != 1 {
		return false, "", nil
	}
	if time.Now().After(lock.ExpireTime) {
		return false, "", nil
	}
	return true, lock.LockHolder, nil
}

func (l *DistributedLock) GetLock(ctx context.Context, key string) (*LockInfo, error) {
	lock, err := l.q.GetLockInfoAll(ctx, key)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get lock: %w", err)
	}
	if lock.LockKey == "" {
		return nil, nil
	}
	return convertLockInfo(lock), nil
}

func (l *DistributedLock) ForceRelease(ctx context.Context, key string) error {
	_, err := l.q.ForceReleaseLock(ctx, key)
	return err
}

func (l *DistributedLock) CountActive(ctx context.Context) (int64, error) {
	return l.q.CountActiveLocks(ctx)
}

func (l *DistributedLock) ListActive(ctx context.Context, limit, offset int32) ([]LockInfo, error) {
	locks, err := l.q.ListActiveLocks(ctx, repo.ListActiveLocksParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	result := make([]LockInfo, 0, len(locks))
	for _, lock := range locks {
		result = append(result, *convertLockInfo(lock))
	}
	return result, nil
}

func (l *DistributedLock) CleanExpired(ctx context.Context) (int64, error) {
	result, err := l.q.CleanExpiredLocks(ctx)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (l *DistributedLock) StartCleaner(parent context.Context, interval time.Duration) context.CancelFunc {
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	ctx, cancel := context.WithCancel(parent)
	registerCleanup(cancel)

	go func() {
		logger := GetLogger()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logger.Info("[distributed_lock] cleaner stopped")
				return
			case <-ticker.C:
				n, err := l.CleanExpired(ctx)
				if err != nil {
					logger.Warnf("[distributed_lock] cleanup failed: %v", err)
				} else if n > 0 {
					logger.Infof("[distributed_lock] cleaned %d expired locks", n)
				}
			}
		}
	}()

	logger := GetLogger()
	logger.Infof("[distributed_lock] cleaner started (interval=%v)", interval)
	return cancel
}

func convertLockInfo(l repo.SysDistributedLock) *LockInfo {
	return &LockInfo{
		ID:          l.ID,
		GmtCreate:   l.GmtCreate,
		GmtModified: l.GmtModified,
		LockKey:     l.LockKey,
		LockHolder:  l.LockHolder,
		LockTTL:     int(l.LockTtl),
		ExpireTime:  l.ExpireTime,
		IsActive:    l.IsActive == 1,
	}
}
