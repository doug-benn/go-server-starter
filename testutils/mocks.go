package testutils

import (
	"context"

	"github.com/doug-benn/go-server-starter/models"
)

// MockTodoRepository implements repository. TodoRepository for testing
type MockTodoRepository struct {
	CreateFunc          func(ctx context.Context, todo *models.Todo) error
	GetByIDFunc         func(ctx context.Context, id int64) (*models.Todo, error)
	GetAllFunc          func(ctx context.Context) (models.Todos, error)
	UpdateFunc          func(ctx context.Context, todo *models.Todo) error
	DeleteFunc          func(ctx context.Context, id int64) error
	MarkAsCompletedFunc func(ctx context.Context, id int64) error
}

func (m *MockTodoRepository) Create(ctx context.Context, todo *models.Todo) error {
	return m.CreateFunc(ctx, todo)
}

func (m *MockTodoRepository) GetByID(ctx context.Context, id int64) (*models.Todo, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *MockTodoRepository) GetAll(ctx context.Context) (models.Todos, error) {
	return m.GetAllFunc(ctx)
}

func (m *MockTodoRepository) Update(ctx context.Context, todo *models.Todo) error {
	return m.UpdateFunc(ctx, todo)
}

func (m *MockTodoRepository) Delete(ctx context.Context, id int64) error {
	return m.DeleteFunc(ctx, id)
}

func (m *MockTodoRepository) MarkAsCompleted(ctx context.Context, id int64) error {
	return m.MarkAsCompletedFunc(ctx, id)
}

// MockTodoService implements service.TodoService for testing
type MockTodoService struct {
	CreateTodoFunc   func(ctx context.Context, title, description string) (*models.Todo, error)
	GetTodoByIDFunc  func(ctx context.Context, id int64) (*models.Todo, error)
	GetAllTodosFunc  func(ctx context.Context) (models.Todos, error)
	UpdateTodoFunc   func(ctx context.Context, todo *models.Todo) error
	DeleteTodoFunc   func(ctx context.Context, id int64) error
	CompleteTodoFunc func(ctx context.Context, id int64) error
}

func (m *MockTodoService) CreateTodo(ctx context.Context, title, description string) (*models.Todo, error) {
	return m.CreateTodoFunc(ctx, title, description)
}

func (m *MockTodoService) GetTodoByID(ctx context.Context, id int64) (*models.Todo, error) {
	return m.GetTodoByIDFunc(ctx, id)
}

func (m *MockTodoService) GetAllTodos(ctx context.Context) (models.Todos, error) {
	return m.GetAllTodosFunc(ctx)
}

func (m *MockTodoService) UpdateTodo(ctx context.Context, todo *models.Todo) error {
	return m.UpdateTodoFunc(ctx, todo)
}

func (m *MockTodoService) DeleteTodo(ctx context.Context, id int64) error {
	return m.DeleteTodoFunc(ctx, id)
}

func (m *MockTodoService) CompleteTodo(ctx context.Context, id int64) error {
	return m.CompleteTodoFunc(ctx, id)
}
