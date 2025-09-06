package todo_list

import (
	"context"

	"assistant/internal/app/repo"
	"assistant/internal/db"
)

type TodoListService struct {
	q *repo.Queries
}

func NewTodoListService() *TodoListService {
	return &TodoListService{
		q: repo.New(db.GetDB()),
	}
}

func (s *TodoListService) CreateTodoList(ctx context.Context, arg repo.CreateTodoListParams) error {
	_, err := s.q.CreateTodoList(ctx, arg)
	return err
}

func (s *TodoListService) GetTodoList(ctx context.Context, id int64) (repo.TodoList, error) {
	return s.q.GetTodoList(ctx, id)
}

func (s *TodoListService) DoneTodoList(ctx context.Context, id int64) error {
	_, err := s.q.DoneTodoList(ctx, id)
	return err
}

func (s *TodoListService) DeleteTodoList(ctx context.Context, id int64) error {
	_, err := s.q.DeleteTodoList(ctx, id)
	return err
}
