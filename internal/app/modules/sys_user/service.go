package sys_user

import (
	"context"
	"database/sql"

	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
)

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
