package sys_user

import (
	"context"
	"database/sql"
	"errors"

	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
	"assistant/pkg/hash"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

type SysUserService struct {
	q *repo.Queries
}

func NewSysUserService() *SysUserService {
	return &SysUserService{
		q: repo.New(psl.GetDB()),
	}
}

func (s *SysUserService) CountSysUsers(ctx context.Context) (int64, error) {
	return s.q.CountSysUsers(ctx)
}

func (s *SysUserService) CreateSysUser(ctx context.Context, arg repo.CreateSysUserParams) (sql.Result, error) {
	return s.q.CreateSysUser(ctx, arg)
}

func (s *SysUserService) DeleteSysUserByID(ctx context.Context, id int64) (sql.Result, error) {
	return s.q.DeleteSysUserByID(ctx, id)
}

func (s *SysUserService) GetSysUserByID(ctx context.Context, id int64) (repo.SysUser, error) {
	return s.q.GetSysUserByID(ctx, id)
}

func (s *SysUserService) ListSysUsers(ctx context.Context, arg repo.ListSysUsersParams) ([]repo.SysUser, error) {
	return s.q.ListSysUsers(ctx, arg)
}

func (s *SysUserService) SearchSysUsersByEmail(ctx context.Context, arg repo.SearchSysUsersByEmailParams) ([]repo.SysUser, error) {
	return s.q.SearchSysUsersByEmail(ctx, arg)
}

func (s *SysUserService) SearchSysUsersByUserName(ctx context.Context, arg repo.SearchSysUsersByUserNameParams) ([]repo.SysUser, error) {
	return s.q.SearchSysUsersByUserName(ctx, arg)
}

func (s *SysUserService) UpdateSysUserByID(ctx context.Context, arg repo.UpdateSysUserByIDParams) (sql.Result, error) {
	return s.q.UpdateSysUserByID(ctx, arg)
}

func (s *SysUserService) Login(ctx context.Context, username, password string) (repo.SysUser, error) {
	users, err := s.q.SearchSysUsersByUserName(ctx, repo.SearchSysUsersByUserNameParams{
		UserName: username,
		Limit:    1,
		Offset:   0,
	})
	if err != nil {
		return repo.SysUser{}, err
	}
	if len(users) == 0 {
		return repo.SysUser{}, ErrInvalidCredentials
	}
	user := users[0]
	if err := hash.CheckPasswordHash(user.Password, password); err != nil {
		return repo.SysUser{}, ErrInvalidCredentials
	}
	return user, nil
}

func (s *SysUserService) DebugListUsers(ctx context.Context) ([]repo.SysUser, error) {
	return s.q.ListSysUsers(ctx, repo.ListSysUsersParams{Limit: 10, Offset: 0})
}
