package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/doug-benn/go-server-starter/models"
	"github.com/doug-benn/go-server-starter/repository"
)

type TodoService interface {
	CreateTodo(ctx context.Context, title, description string) (*models.Todo, error)
	GetTodoByID(ctx context.Context, id int32) (*models.Todo, error)
	GetAllTodos(ctx context.Context) ([]models.Todo, error)
	UpdateTodo(ctx context.Context, todo *models.Todo) error
	DeleteTodo(ctx context.Context, id int32) error
	CompleteTodo(ctx context.Context, id int32) error
}

type TodoServiceImpl struct {
	repo   repository.Querier
	logger *slog.Logger
}

func NewTodoService(repo repository.Querier, logger *slog.Logger) TodoService {
	return &TodoServiceImpl{repo: repo, logger: logger}
}

func (s *TodoServiceImpl) CreateTodo(ctx context.Context, title, description string) (*models.Todo, error) {
	now := time.Now()
	todo, err := s.repo.CreateTodo(ctx, repository.CreateTodoParams{
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create todo", "error", err)
		return nil, err
	}
	return &todo, nil
}

func (s *TodoServiceImpl) GetTodoByID(ctx context.Context, id int32) (*models.Todo, error) {
	todo, err := s.repo.GetTodo(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get todo by id", "id", id, "error", err)
		return nil, err
	}
	return &todo, nil
}

func (s *TodoServiceImpl) GetAllTodos(ctx context.Context) ([]models.Todo, error) {
	todos, err := s.repo.ListTodos(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list todos", "error", err)
		return nil, err
	}
	return todos, nil
}

func (s *TodoServiceImpl) UpdateTodo(ctx context.Context, todo *models.Todo) error {
	updated, err := s.repo.UpdateTodo(ctx, repository.UpdateTodoParams{
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
		UpdatedAt:   time.Now(),
		ID:          todo.ID,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to update todo", "id", todo.ID, "error", err)
		return err
	}
	*todo = updated
	return nil
}

func (s *TodoServiceImpl) DeleteTodo(ctx context.Context, id int32) error {
	err := s.repo.DeleteTodo(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to delete todo", "id", id, "error", err)
		return err
	}
	return nil
}

func (s *TodoServiceImpl) CompleteTodo(ctx context.Context, id int32) error {
	_, err := s.repo.CompleteTodo(ctx, repository.CompleteTodoParams{
		UpdatedAt: time.Now(),
		ID:        id,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to complete todo", "id", id, "error", err)
		return err
	}
	return nil
}
