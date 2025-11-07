package task_instance

import (
	"context"
	"database/sql"

	"assistant/internal/app/repo"
	"assistant/internal/db"
)

type TaskInstanceService struct {
	q *repo.Queries
}

func NewTaskInstanceService() *TaskInstanceService {
	return &TaskInstanceService{
		q: repo.New(db.GetDB()),
	}
}

func (s *TaskInstanceService) CountTaskInstances(ctx context.Context) (int64, error) {
	return s.q.CountTaskInstances(ctx)
}

func (s *TaskInstanceService) CreateTaskInstance(ctx context.Context, arg repo.CreateTaskInstanceParams) error {
	_, err := s.q.CreateTaskInstance(ctx, arg)
	return err
}

func (s *TaskInstanceService) DeleteTaskInstanceByID(ctx context.Context, id int64) (sql.Result, error) {
	return s.q.DeleteTaskInstanceByID(ctx, id)
}

func (s *TaskInstanceService) GetTaskInstanceByID(ctx context.Context, id int64) (repo.TaskInstance, error) {
	return s.q.GetTaskInstanceByID(ctx, id)
}

func (s *TaskInstanceService) ListTaskInstances(ctx context.Context, arg repo.ListTaskInstancesParams) ([]repo.TaskInstance, error) {
	return s.q.ListTaskInstances(ctx, arg)
}

func (s *TaskInstanceService) UpdateTaskInstanceByID(ctx context.Context, arg repo.UpdateTaskInstanceByIDParams) (sql.Result, error) {
	return s.q.UpdateTaskInstanceByID(ctx, arg)
}
