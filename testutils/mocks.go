package testutils

import (
	"context"

	"github.com/doug-benn/go-server-starter/models"
	"github.com/doug-benn/go-server-starter/repository"
)

type MockQuerier struct {
	CreateTodoFunc   func(ctx context.Context, arg repository.CreateTodoParams) (models.Todo, error)
	GetTodoFunc      func(ctx context.Context, id int32) (models.Todo, error)
	ListTodosFunc    func(ctx context.Context) ([]models.Todo, error)
	UpdateTodoFunc   func(ctx context.Context, arg repository.UpdateTodoParams) (models.Todo, error)
	DeleteTodoFunc   func(ctx context.Context, id int32) error
	CompleteTodoFunc func(ctx context.Context, arg repository.CompleteTodoParams) (models.Todo, error)
}

func (m *MockQuerier) CreateTodo(ctx context.Context, arg repository.CreateTodoParams) (models.Todo, error) {
	return m.CreateTodoFunc(ctx, arg)
}

func (m *MockQuerier) GetTodo(ctx context.Context, id int32) (models.Todo, error) {
	return m.GetTodoFunc(ctx, id)
}

func (m *MockQuerier) ListTodos(ctx context.Context) ([]models.Todo, error) {
	return m.ListTodosFunc(ctx)
}

func (m *MockQuerier) UpdateTodo(ctx context.Context, arg repository.UpdateTodoParams) (models.Todo, error) {
	return m.UpdateTodoFunc(ctx, arg)
}

func (m *MockQuerier) DeleteTodo(ctx context.Context, id int32) error {
	return m.DeleteTodoFunc(ctx, id)
}

func (m *MockQuerier) CompleteTodo(ctx context.Context, arg repository.CompleteTodoParams) (models.Todo, error) {
	return m.CompleteTodoFunc(ctx, arg)
}

var _ repository.Querier = (*MockQuerier)(nil)
