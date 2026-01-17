package health

import (
	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
)

type HealthService struct {
	q *repo.Queries
}

func NewHealthService() *HealthService {
	return &HealthService{
		q: repo.New(psl.GetDB()),
	}
}

func (s *HealthService) Health() (string, error) {
	// TODO: <23:09:23 2025-10-23: Dionysus>:
	return "ok", nil
}
