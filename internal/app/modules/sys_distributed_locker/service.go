package sys_distributed_locker

import (
	"context"
	"fmt"

	"assistant/internal/bootstrap/psl"
)

type SysDistributedLockService struct {
	locker *psl.DistributedLock
}

func NewSysDistributedLockService() *SysDistributedLockService {
	return &SysDistributedLockService{
		locker: psl.GetDistributedLock(),
	}
}

func (s *SysDistributedLockService) audit(op, key, holder string, success bool, reason string) {
	logger := psl.GetLogger()
	auditor := "system"
	if holder != "" && len(holder) > 8 {
		auditor = holder[:8] + "..."
	}
	result := "success"
	if !success {
		result = "failed"
	}
	msg := fmt.Sprintf("[AUDIT] dislock op=%s key=%s holder=%s result=%s", op, key, auditor, result)
	if reason != "" {
		msg += fmt.Sprintf(" reason=%s", reason)
	}
	logger.Info(msg)
}

func (s *SysDistributedLockService) TryAcquire(ctx context.Context, key, holder string, ttl int) (bool, error) {
	if s.locker == nil {
		s.audit("acquire", key, holder, false, "dislocker not initialized")
		return false, fmt.Errorf("dislocker not initialized")
	}
	success, err := s.locker.TryAcquire(ctx, key, holder, ttl)
	s.audit("acquire", key, holder, success, "")
	return success, err
}

func (s *SysDistributedLockService) Release(ctx context.Context, key, holder string) (bool, error) {
	if s.locker == nil {
		s.audit("release", key, holder, false, "dislocker not initialized")
		return false, fmt.Errorf("dislocker not initialized")
	}
	success, err := s.locker.Release(ctx, key, holder)
	s.audit("release", key, holder, success, "")
	return success, err
}

func (s *SysDistributedLockService) Renew(ctx context.Context, key, holder string, ttl int) (bool, error) {
	if s.locker == nil {
		s.audit("renew", key, holder, false, "dislocker not initialized")
		return false, fmt.Errorf("dislocker not initialized")
	}
	success, err := s.locker.Renew(ctx, key, holder, ttl)
	s.audit("renew", key, holder, success, "")
	return success, err
}

func (s *SysDistributedLockService) Get(ctx context.Context, key string) (*psl.LockInfo, error) {
	if s.locker == nil {
		return nil, fmt.Errorf("dislocker not initialized")
	}
	info, err := s.locker.GetLock(ctx, key)
	if err == nil {
		logger := psl.GetLogger()
		logger.Infof("[AUDIT] dislock query key=%s holder=%s ttl=%d active=%v",
			key, info.LockHolder, info.LockTTL, info.IsActive)
	}
	return info, err
}

func (s *SysDistributedLockService) IsHeld(ctx context.Context, key string) (bool, string, error) {
	if s.locker == nil {
		return false, "", fmt.Errorf("dislocker not initialized")
	}
	held, holder, err := s.locker.IsHeld(ctx, key)
	logger := psl.GetLogger()
	logger.Infof("[AUDIT] dislock check key=%s held=%v holder=%s", key, held, holder)
	return held, holder, err
}

func (s *SysDistributedLockService) ForceRelease(ctx context.Context, key string) error {
	if s.locker == nil {
		s.audit("force_release", key, "admin", false, "dislocker not initialized")
		return fmt.Errorf("dislocker not initialized")
	}
	err := s.locker.ForceRelease(ctx, key)
	s.audit("force_release", key, "admin", err == nil, "")
	return err
}

func (s *SysDistributedLockService) CountActive(ctx context.Context) (int64, error) {
	if s.locker == nil {
		return 0, fmt.Errorf("dislocker not initialized")
	}
	return s.locker.CountActive(ctx)
}

func (s *SysDistributedLockService) ListActive(ctx context.Context, limit, offset int32) ([]psl.LockInfo, error) {
	if s.locker == nil {
		return nil, fmt.Errorf("dislocker not initialized")
	}
	return s.locker.ListActive(ctx, limit, offset)
}
