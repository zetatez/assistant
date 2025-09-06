package health

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Health() (string, error) {
	// TODO: <16:39:48 2025-09-06: Dionysus>:
	return "ok", nil
}
