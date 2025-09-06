package health

import (
	"context"
	"time"

	"assistant/internal/bootstrap/psl"
)

type HealthStatus struct {
	Status   string           `json:"status"`
	Duration string           `json:"duration"`
	Checks   map[string]Check `json:"checks"`
}

type Check struct {
	Status   string `json:"status"`
	Duration string `json:"duration"`
	Error    string `json:"error,omitempty"`
}

func NewHealthService() *HealthService {
	return &HealthService{}
}

type HealthService struct{}

func (s *HealthService) Health() (*HealthStatus, error) {
	status := &HealthStatus{
		Status: "ok",
		Checks: make(map[string]Check),
	}
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbCheck := s.checkDatabase(ctx)
	status.Checks["database"] = dbCheck
	if dbCheck.Status != "ok" {
		status.Status = "degraded"
	}

	dislockCheck := s.checkDisLocker(ctx)
	status.Checks["distributed_locker"] = dislockCheck
	if dislockCheck.Status != "ok" {
		status.Status = "degraded"
	}

	status.Duration = time.Since(start).String()
	return status, nil
}

func (s *HealthService) checkDatabase(ctx context.Context) Check {
	start := time.Now()
	db := psl.GetDB()
	if db == nil {
		return Check{
			Status:   "error",
			Duration: time.Since(start).String(),
			Error:    "database not initialized",
		}
	}
	if err := db.PingContext(ctx); err != nil {
		psl.GetLogger().Errorf("database ping failed: %v", err)
		return Check{
			Status:   "error",
			Duration: time.Since(start).String(),
			Error:    "database ping failed",
		}
	}
	return Check{
		Status:   "ok",
		Duration: time.Since(start).String(),
	}
}

func (s *HealthService) checkDisLocker(ctx context.Context) Check {
	start := time.Now()
	locker := psl.GetDistributedLock()
	if locker == nil {
		return Check{
			Status:   "error",
			Duration: time.Since(start).String(),
			Error:    "dislocker not initialized",
		}
	}
	_, err := locker.CountActive(ctx)
	if err != nil {
		psl.GetLogger().Errorf("dislocker check failed: %v", err)
		return Check{
			Status:   "error",
			Duration: time.Since(start).String(),
			Error:    "dislocker check failed",
		}
	}
	return Check{
		Status:   "ok",
		Duration: time.Since(start).String(),
	}
}
