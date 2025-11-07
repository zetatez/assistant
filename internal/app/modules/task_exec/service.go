package task_exec

import (
	"context"
	"database/sql"

	"assistant/internal/app/repo"
	"assistant/internal/db"
)

type TaskExecService struct {
	q *repo.Queries
}

func NewTaskExecService() *TaskExecService {
	return &TaskExecService{
		q: repo.New(db.GetDB()),
	}
}

func (s *TaskExecService) CountTaskExecs(ctx context.Context) (int64, error) {
	return s.q.CountTaskExecs(ctx)
}

func (s *TaskExecService) CreateTaskExec(ctx context.Context, arg repo.CreateTaskExecParams) error {
	_, err := s.q.CreateTaskExec(ctx, arg)
	return err
}

func (s *TaskExecService) DeleteTaskExecByID(ctx context.Context, id int64) (sql.Result, error) {
	return s.q.DeleteTaskExecByID(ctx, id)
}

func (s *TaskExecService) GetTaskExecByID(ctx context.Context, id int64) (repo.TaskExec, error) {
	return s.q.GetTaskExecByID(ctx, id)
}

func (s *TaskExecService) ListTaskExecs(ctx context.Context, arg repo.ListTaskExecsParams) ([]repo.TaskExec, error) {
	return s.q.ListTaskExecs(ctx, arg)
}

func (s *TaskExecService) UpdateTaskExecByID(ctx context.Context, arg repo.UpdateTaskExecByIDParams) (sql.Result, error) {
	return s.q.UpdateTaskExecByID(ctx, arg)
}
