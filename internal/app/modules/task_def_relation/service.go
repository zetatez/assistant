package task_def_relation

import (
	"context"
	"database/sql"

	"assistant/internal/app/repo"
	"assistant/internal/db"
)

type TaskDefRelationService struct {
	q *repo.Queries
}

func NewTaskDefRelationService() *TaskDefRelationService {
	return &TaskDefRelationService{
		q: repo.New(db.GetDB()),
	}
}

func (s *TaskDefRelationService) CountTaskDefRelations(ctx context.Context) (int64, error) {
	return s.q.CountTaskDefRelations(ctx)
}

func (s *TaskDefRelationService) CreateTaskDefRelation(ctx context.Context, arg repo.CreateTaskDefRelationParams) error {
	_, err := s.q.CreateTaskDefRelation(ctx, arg)
	return err
}

func (s *TaskDefRelationService) DeleteTaskDefRelationByID(ctx context.Context, id int64) (sql.Result, error) {
	return s.q.DeleteTaskDefRelationByID(ctx, id)
}

func (s *TaskDefRelationService) GetTaskDefRelationByID(ctx context.Context, id int64) (repo.TaskDefRelation, error) {
	return s.q.GetTaskDefRelationByID(ctx, id)
}

func (s *TaskDefRelationService) ListTaskDefRelations(ctx context.Context, arg repo.ListTaskDefRelationsParams) ([]repo.TaskDefRelation, error) {
	return s.q.ListTaskDefRelations(ctx, arg)
}

func (s *TaskDefRelationService) UpdateTaskDefRelationByID(ctx context.Context, arg repo.UpdateTaskDefRelationByIDParams) (sql.Result, error) {
	return s.q.UpdateTaskDefRelationByID(ctx, arg)
}
