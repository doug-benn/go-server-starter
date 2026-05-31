package services_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/doug-benn/go-server-starter/models"
	"github.com/doug-benn/go-server-starter/repository"
	"github.com/doug-benn/go-server-starter/services"
	"github.com/doug-benn/go-server-starter/testutils"
)

func now() time.Time {
	return time.Now().Truncate(time.Microsecond)
}

func TestCreateTodo(t *testing.T) {
	mockRepo := &testutils.MockQuerier{
		CreateTodoFunc: func(ctx context.Context, arg repository.CreateTodoParams) (models.Todo, error) {
			return models.Todo{
				ID:          1,
				Title:       arg.Title,
				Description: arg.Description,
				Completed:   arg.Completed,
				CreatedAt:   arg.CreatedAt,
				UpdatedAt:   arg.UpdatedAt,
			}, nil
		},
	}

	todoService := services.NewTodoService(mockRepo, slog.Default())
	ctx := context.Background()

	todo, err := todoService.CreateTodo(ctx, "Test Todo", "Test Description")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if todo == nil {
		t.Fatal("Expected todo to be created, got nil")
	}
	if todo.ID != 1 {
		t.Errorf("Expected ID 1, got %d", todo.ID)
	}
	if todo.Title != "Test Todo" {
		t.Errorf("Expected title 'Test Todo', got '%s'", todo.Title)
	}
	if todo.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", todo.Description)
	}
	if todo.Completed {
		t.Errorf("Expected completed to be false, got true")
	}
}

func TestCreateTodo_Error(t *testing.T) {
	expectedErr := errors.New("database error")
	mockRepo := &testutils.MockQuerier{
		CreateTodoFunc: func(ctx context.Context, arg repository.CreateTodoParams) (models.Todo, error) {
			return models.Todo{}, expectedErr
		},
	}

	todoService := services.NewTodoService(mockRepo, slog.Default())
	ctx := context.Background()

	todo, err := todoService.CreateTodo(ctx, "Test Todo", "Test Description")

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if todo != nil {
		t.Errorf("Expected todo to be nil, got %v", todo)
	}
}

func TestGetTodoByID(t *testing.T) {
	ts := now()
	expectedTodo := models.Todo{
		ID:          1,
		Title:       "Test Todo",
		Description: "Test Description",
		Completed:   false,
		CreatedAt:   ts,
		UpdatedAt:   ts,
	}

	mockRepo := &testutils.MockQuerier{
		GetTodoFunc: func(ctx context.Context, id int32) (models.Todo, error) {
			if id == 1 {
				return expectedTodo, nil
			}
			return models.Todo{}, errors.New("not found")
		},
	}

	todoService := services.NewTodoService(mockRepo, slog.Default())
	ctx := context.Background()

	todo, err := todoService.GetTodoByID(ctx, 1)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if todo == nil {
		t.Fatal("Expected todo to be returned, got nil")
	}
	if todo.ID != expectedTodo.ID {
		t.Errorf("Expected ID %d, got %d", expectedTodo.ID, todo.ID)
	}
	if todo.Title != expectedTodo.Title {
		t.Errorf("Expected title '%s', got '%s'", expectedTodo.Title, todo.Title)
	}
}

func TestGetAllTodos(t *testing.T) {
	ts := now()
	expectedTodos := []models.Todo{
		{
			ID:          1,
			Title:       "Todo 1",
			Description: "Description 1",
			Completed:   false,
			CreatedAt:   ts,
			UpdatedAt:   ts,
		},
		{
			ID:          2,
			Title:       "Todo 2",
			Description: "Description 2",
			Completed:   true,
			CreatedAt:   ts,
			UpdatedAt:   ts,
		},
	}

	mockRepo := &testutils.MockQuerier{
		ListTodosFunc: func(ctx context.Context) ([]models.Todo, error) {
			return expectedTodos, nil
		},
	}

	todoService := services.NewTodoService(mockRepo, slog.Default())
	ctx := context.Background()

	todos, err := todoService.GetAllTodos(ctx)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(todos) != len(expectedTodos) {
		t.Fatalf("Expected %d todos, got %d", len(expectedTodos), len(todos))
	}
	for i, todo := range todos {
		if todo.ID != expectedTodos[i].ID {
			t.Errorf("Todo %d: Expected ID %d, got %d", i, expectedTodos[i].ID, todo.ID)
		}
		if todo.Title != expectedTodos[i].Title {
			t.Errorf("Todo %d: Expected title '%s', got '%s'", i, expectedTodos[i].Title, todo.Title)
		}
	}
}

func TestCompleteTodo(t *testing.T) {
	var completedID int32
	mockRepo := &testutils.MockQuerier{
		CompleteTodoFunc: func(ctx context.Context, arg repository.CompleteTodoParams) (models.Todo, error) {
			completedID = arg.ID
			return models.Todo{}, nil
		},
	}

	todoService := services.NewTodoService(mockRepo, slog.Default())
	ctx := context.Background()

	err := todoService.CompleteTodo(ctx, 1)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if completedID != 1 {
		t.Errorf("Expected to complete todo with ID 1, got %d", completedID)
	}
}
