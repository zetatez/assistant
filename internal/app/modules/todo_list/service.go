package todo_list

import (
	"context"
	"database/sql"

	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"
)

type TodoListService struct {
	q *repo.Queries
}

func NewTodoListService() *TodoListService {
	return &TodoListService{
		q: repo.New(psl.GetDB()),
	}
}

func (s *TodoListService) CountTodoList(ctx context.Context) (int64, error) {
	return s.q.CountTodoList(ctx)
}

func (s *TodoListService) CreateTodoList(ctx context.Context, arg repo.CreateTodoListParams) error {
	_, err := s.q.CreateTodoList(ctx, arg)
	return err
}

func (s *TodoListService) DeleteTodoListByID(ctx context.Context, id int64) error {
	_, err := s.q.DeleteTodoListByID(ctx, id)
	return err
}

func (s *TodoListService) GetTodoListByID(ctx context.Context, id int64) (repo.TodoList, error) {
	return s.q.GetTodoListByID(ctx, id)
}

func (s *TodoListService) ListTodoLists(ctx context.Context, arg repo.ListTodoListsParams) ([]repo.TodoList, error) {
	return s.q.ListTodoLists(ctx, arg)
}

func (s *TodoListService) SearchTodoListsByContent(ctx context.Context, arg repo.SearchTodoListsByContentParams) ([]repo.TodoList, error) {
	return s.q.SearchTodoListsByContent(ctx, arg)
}

func (s *TodoListService) SearchTodoListsByDeadlineLT(ctx context.Context, arg repo.SearchTodoListsByDeadlineLTParams) ([]repo.TodoList, error) {
	return s.q.SearchTodoListsByDeadlineLT(ctx, arg)
}

func (s *TodoListService) SearchTodoListsByTitle(ctx context.Context, arg repo.SearchTodoListsByTitleParams) ([]repo.TodoList, error) {
	return s.q.SearchTodoListsByTitle(ctx, arg)
}

func (s *TodoListService) SearchTodoListsByTitleAndContent(ctx context.Context, arg repo.SearchTodoListsByTitleAndContentParams) ([]repo.TodoList, error) {
	return s.q.SearchTodoListsByTitleAndContent(ctx, arg)
}

func (s *TodoListService) UpdateTodoListByID(ctx context.Context, arg repo.UpdateTodoListByIDParams) (sql.Result, error) {
	return s.q.UpdateTodoListByID(ctx, arg)
}

func (s *TodoListService) UpdateTodoListProgressByID(ctx context.Context, id int64, progress int64) error {
	progress32 := int32(progress)
	_, err := s.q.UpdateTodoListProgressByID(ctx, repo.UpdateTodoListProgressByIDParams{
		Progress: progress32,
		Column2:  progress32,
		ID:       id,
	})
	return err
}

func (s *TodoListService) CompleteTodoListByID(ctx context.Context, id int64) error {
	_, err := s.q.CompleteTodoListByID(ctx, id)
	return err
}

func (s *TodoListService) UpdateTodoListPriorityByID(ctx context.Context, id int64, priority int64) error {
	_, err := s.q.UpdateTodoListPriorityByID(ctx, repo.UpdateTodoListPriorityByIDParams{
		Priority: int32(priority),
		ID:       id,
	})
	return err
}
