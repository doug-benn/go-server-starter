package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/doug-benn/go-server-starter/models"
	"github.com/doug-benn/go-server-starter/services"
	"github.com/doug-benn/go-server-starter/testutils"
)

func TestCreateTodo(t *testing.T) {
	// Arrange
	mockRepo := &testutils.MockTodoRepository{
		CreateFunc: func(ctx context.Context, todo *models.Todo) error {
			// Simulate successful creation by setting an ID
			todo.ID = 1
			todo.CreatedAt = time.Now()
			todo.UpdatedAt = time.Now()
			return nil
		},
	}

	todoService := services.NewTodoService(mockRepo)
	ctx := context.Background()

	// Act
	todo, err := todoService.CreateTodo(ctx, "Test Todo", "Test Description")

	// Assert
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
		t.Error("Expected completed to be false")
	}
}

func TestCreateTodo_Error(t *testing.T) {
	// Arrange
	expectedErr := errors.New("database error")
	mockRepo := &testutils.MockTodoRepository{
		CreateFunc: func(ctx context.Context, todo *models.Todo) error {
			return expectedErr
		},
	}

	todoService := services.NewTodoService(mockRepo)
	ctx := context.Background()

	// Act
	todo, err := todoService.CreateTodo(ctx, "Test Todo", "Test Description")

	// Assert
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if todo != nil {
		t.Errorf("Expected todo to be nil, got %v", todo)
	}
}

func TestGetTodoByID(t *testing.T) {
	// Arrange
	now := time.Now()
	expectedTodo := &models.Todo{
		ID:          1,
		Title:       "Test Todo",
		Description: "Test Description",
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockRepo := &testutils.MockTodoRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.Todo, error) {
			if id == 1 {
				return expectedTodo, nil
			}
			return nil, nil
		},
	}

	todoService := services.NewTodoService(mockRepo)
	ctx := context.Background()

	// Act
	todo, err := todoService.GetTodoByID(ctx, 1)

	// Assert
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
	// Arrange
	now := time.Now()
	expectedTodos := models.Todos{
		{
			ID:          1,
			Title:       "Todo 1",
			Description: "Description 1",
			Completed:   false,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2,
			Title:       "Todo 2",
			Description: "Description 2",
			Completed:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	mockRepo := &testutils.MockTodoRepository{
		GetAllFunc: func(ctx context.Context) (models.Todos, error) {
			return expectedTodos, nil
		},
	}

	todoService := services.NewTodoService(mockRepo)
	ctx := context.Background()

	// Act
	todos, err := todoService.GetAllTodos(ctx)

	// Assert
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
	// Arrange
	var completedID int64
	mockRepo := &testutils.MockTodoRepository{
		MarkAsCompletedFunc: func(ctx context.Context, id int64) error {
			completedID = id
			return nil
		},
	}

	todoService := services.NewTodoService(mockRepo)
	ctx := context.Background()

	// Act
	err := todoService.CompleteTodo(ctx, 1)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if completedID != 1 {
		t.Errorf("Expected to complete todo with ID 1, got %d", completedID)
	}
}
