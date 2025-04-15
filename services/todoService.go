package services

import (
	"context"

	"github.com/doug-benn/go-server-starter/models"
	"github.com/doug-benn/go-server-starter/repository"
)

type TodoService interface {
	CreateTodo(ctx context.Context, title, description string) (*models.Todo, error)
	GetTodoByID(ctx context.Context, id int64) (*models.Todo, error)
	GetAllTodos(ctx context.Context) (models.Todos, error)
	UpdateTodo(ctx context.Context, todo *models.Todo) error
	DeleteTodo(ctx context.Context, id int64) error
	CompleteTodo(ctx context.Context, id int64) error
}

// With a concrete implementation
type TodoServiceImpl struct {
	repo repository.TodoRepository
}

func NewTodoService(repo repository.TodoRepository) TodoService {
	return &TodoServiceImpl{
		repo: repo,
	}
}

func (s *TodoServiceImpl) CreateTodo(ctx context.Context, title, description string) (*models.Todo, error) {
	todo := &models.Todo{
		Title:       title,
		Description: description,
		Completed:   false,
	}

	if err := s.repo.Create(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

func (s *TodoServiceImpl) GetTodoByID(ctx context.Context, id int64) (*models.Todo, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TodoServiceImpl) GetAllTodos(ctx context.Context) (models.Todos, error) {
	return s.repo.GetAll(ctx)
}

func (s *TodoServiceImpl) UpdateTodo(ctx context.Context, todo *models.Todo) error {
	return s.repo.Update(ctx, todo)
}

func (s *TodoServiceImpl) DeleteTodo(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *TodoServiceImpl) CompleteTodo(ctx context.Context, id int64) error {
	return s.repo.MarkAsCompleted(ctx, id)
}
