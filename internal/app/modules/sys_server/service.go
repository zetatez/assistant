package sys_server

import (
	"context"

	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
)

type SysServerService struct {
	q *repo.Queries
}

func NewSysServerService() *SysServerService {
	return &SysServerService{q: repo.New(psl.GetDB())}
}

func (s *SysServerService) CountSysServers(ctx context.Context) (int64, error) {
	return s.q.CountSysServers(ctx)
}

func (s *SysServerService) GetSysServerByID(ctx context.Context, id int64) (repo.SysServer, error) {
	return s.q.GetSysServerByID(ctx, id)
}

func (s *SysServerService) ListSysServers(ctx context.Context, arg repo.ListSysServersParams) ([]repo.SysServer, error) {
	return s.q.ListSysServers(ctx, arg)
}

func (s *SysServerService) SearchSysServersByIDC(ctx context.Context, arg repo.SearchSysServersByIDCParams) ([]repo.SysServer, error) {
	return s.q.SearchSysServersByIDC(ctx, arg)
}

func (s *SysServerService) SearchSysServersBySvrIP(ctx context.Context, arg repo.SearchSysServersBySvrIPParams) ([]repo.SysServer, error) {
	return s.q.SearchSysServersBySvrIP(ctx, arg)
}

func (s *SysServerService) SearchSysServersByIDCAndSvrIP(ctx context.Context, arg repo.SearchSysServersByIDCAndSvrIPParams) ([]repo.SysServer, error) {
	return s.q.SearchSysServersByIDCAndSvrIP(ctx, arg)
}
