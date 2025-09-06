package user

import "assistant/internal/model"

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListUsers() ([]model.User, error) {
	return s.repo.FindAll()
}

func (s *Service) CreateUser(name string) (model.User, error) {
	return s.repo.Create(name)
}
