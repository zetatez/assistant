package service

import (
	"github.com/zetatez/assistant/internal/models"
	"github.com/zetatez/assistant/internal/repository"
)

type UserService interface {
	Create(name, email string) (*models.User, error)
	Get(id uint) (*models.User, error)
	List(page, size int) ([]models.User, int64, error)
	Update(id uint, name, email string) error
	Delete(id uint) error
}

type userService struct{ repo repository.UserRepository }

func NewUserService(r repository.UserRepository) UserService { return &userService{repo: r} }

func (s *userService) Create(name, email string) (*models.User, error) {
	u := &models.User{Name: name, Email: email}
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *userService) Get(id uint) (*models.User, error) { return s.repo.GetByID(id) }

func (s *userService) List(page, size int) ([]models.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size
	return s.repo.List(offset, size)
}

func (s *userService) Update(id uint, name, email string) error {
	return s.repo.Update(&models.User{ID: id, Name: name, Email: email})
}

func (s *userService) Delete(id uint) error { return s.repo.Delete(id) }
