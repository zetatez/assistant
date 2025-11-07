package user

import (
	"context"
	"database/sql"

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

func (s *UserService) CreateUser(ctx context.Context, arg repo.CreateUserParams) (sql.Result, error) {
	return s.q.CreateUser(ctx, arg)
}

func (s *UserService) DeleteUserByID(ctx context.Context, id int64) (sql.Result, error) {
	return s.q.DeleteUserByID(ctx, id)
}

func (s *UserService) GetUserByID(ctx context.Context, id int64) (repo.User, error) {
	return s.q.GetUserByID(ctx, id)
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

func (s *UserService) UpdateUserByID(ctx context.Context, arg repo.UpdateUserByIDParams) (sql.Result, error) {
	return s.q.UpdateUserByID(ctx, arg)
}
