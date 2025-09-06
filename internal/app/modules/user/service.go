package user

import (
	"context"

	"assistant/internal/app/repo"
	"assistant/internal/db"
)

type UserService struct {
	q *repo.Queries
}

func NewUserService() *UserService {
	return &UserService{
		q: repo.New(db.GetDB()),
	}
}

func (s *UserService) CountUsers(ctx context.Context) (int64, error) {
	return s.q.CountServer(ctx)
}

func (s *UserService) CreateUser(ctx context.Context, arg repo.CreateUserParams) error {
	_, err := s.q.CreateUser(ctx, arg)
	return err
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	return s.q.DeleteUser(ctx, id)
}

func (s *UserService) GetUser(ctx context.Context, id int64) (repo.User, error) {
	return s.q.GetUser(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context, arg repo.ListUsersParams) ([]repo.User, error) {
	return s.q.ListUsers(ctx, arg)
}

func (s *UserService) SearchUsersByEmail(ctx context.Context, arg repo.SearchUsersByEmailParams) ([]repo.User, error) {
	return s.q.SearchUsersByEmail(ctx, arg)
}

func (s *UserService) SearchUsersByUserName(ctx context.Context, arg repo.SearchUsersByUserNameParams) ([]repo.User, error) {
	return s.q.SearchUsersByUserName(ctx, arg)
}

func (s *UserService) UpdateUser(ctx context.Context, arg repo.UpdateUserParams) error {
	return s.q.UpdateUser(ctx, arg)
}
