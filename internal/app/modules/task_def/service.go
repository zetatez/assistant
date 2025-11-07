package task_def

import (
	"context"
	"database/sql"

	"assistant/internal/app/repo"
	"assistant/internal/db"
)

type TaskDefService struct {
	q *repo.Queries
}

func NewTaskDefService() *TaskDefService {
	return &TaskDefService{
		q: repo.New(db.GetDB()),
	}
}

func (s *TaskDefService) CountTaskDefs(ctx context.Context) (int64, error) {
	return s.q.CountTaskDefs(ctx)
}

func (s *TaskDefService) CreateTaskDef(ctx context.Context, arg repo.CreateTaskDefParams) error {
	_, err := s.q.CreateTaskDef(ctx, arg)
	return err
}

func (s *TaskDefService) DeleteTaskDefByID(ctx context.Context, id int64) (sql.Result, error) {
	return s.q.DeleteTaskDefByID(ctx, id)
}

func (s *TaskDefService) GetTaskDefByID(ctx context.Context, id int64) (repo.TaskDef, error) {
	return s.q.GetTaskDefByID(ctx, id)
}

func (s *TaskDefService) ListTaskDefs(ctx context.Context, arg repo.ListTaskDefsParams) ([]repo.TaskDef, error) {
	return s.q.ListTaskDefs(ctx, arg)
}

func (s *TaskDefService) UpdateTaskDefByID(ctx context.Context, arg repo.UpdateTaskDefByIDParams) (sql.Result, error) {
	return s.q.UpdateTaskDefByID(ctx, arg)
}
