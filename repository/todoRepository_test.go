package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-benn/go-server-starter/models"
	"github.com/doug-benn/go-server-starter/repository"
)

func TestCreate(t *testing.T) {
	// Create a new mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database connection: %v", err)
	}
	defer db.Close()

	// Create repository with mock db
	repo := repository.NewPostgresTodoRepository(db)

	// Setup expectations
	todo := &models.Todo{
		Title:       "Test Todo",
		Description: "Test Description",
		Completed:   false,
	}

	// We expect a query to be executed with specific parameters
	mock.ExpectQuery("INSERT INTO todos").
		WithArgs(todo.Title, todo.Description, todo.Completed, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Call the method being tested
	ctx := context.Background()
	err = repo.Create(ctx, todo)

	// Verify the results
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if todo.ID != 1 {
		t.Errorf("Expected ID 1, got %d", todo.ID)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestGetByID(t *testing.T) {
	// Create a new mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database connection: %v", err)
	}
	defer db.Close()

	// Create repository with mock db
	repo := repository.NewPostgresTodoRepository(db)

	// Setup expectations
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "title", "description", "completed", "created_at", "updated_at"}).
		AddRow(1, "Test Todo", "Test Description", false, now, now)

	mock.ExpectQuery("SELECT (.+) FROM todos WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(rows)

	// Call the method being tested
	ctx := context.Background()
	todo, err := repo.GetByID(ctx, 1)

	// Verify the results
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if todo == nil {
		t.Fatal("Expected todo to be returned, got nil")
	}
	if todo.ID != 1 {
		t.Errorf("Expected ID 1, got %d", todo.ID)
	}
	if todo.Title != "Test Todo" {
		t.Errorf("Expected title 'Test Todo', got '%s'", todo.Title)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	// Create a new mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database connection: %v", err)
	}
	defer db.Close()

	// Create repository with mock db
	repo := repository.NewPostgresTodoRepository(db)

	// Setup expectations - no rows returned
	mock.ExpectQuery("SELECT (.+) FROM todos WHERE id = \\$1").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	// Call the method being tested
	ctx := context.Background()
	todo, err := repo.GetByID(ctx, 1)

	// Verify the results
	if err != nil {
		t.Errorf("Expected nil error for non-existent todo, got %v", err)
	}
	if todo != nil {
		t.Errorf("Expected nil todo for non-existent ID, got %v", todo)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestGetAll(t *testing.T) {
	// Create a new mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database connection: %v", err)
	}
	defer db.Close()

	// Create repository with mock db
	repo := repository.NewPostgresTodoRepository(db)

	// Setup expectations
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "title", "description", "completed", "created_at", "updated_at"}).
		AddRow(1, "Todo 1", "Description 1", false, now, now).
		AddRow(2, "Todo 2", "Description 2", true, now, now)

	mock.ExpectQuery("SELECT (.+) FROM todos ORDER BY created_at DESC").
		WillReturnRows(rows)

	// Call the method being tested
	ctx := context.Background()
	todos, err := repo.GetAll(ctx)

	// Verify the results
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(todos) != 2 {
		t.Fatalf("Expected 2 todos, got %d", len(todos))
	}
	if todos[0].ID != 1 || todos[1].ID != 2 {
		t.Errorf("Expected todos with IDs 1 and 2, got %d and %d", todos[0].ID, todos[1].ID)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestUpdate(t *testing.T) {
	// Create a new mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database connection: %v", err)
	}
	defer db.Close()

	// Create repository with mock db
	repo := repository.NewPostgresTodoRepository(db)

	// Setup expectations
	todo := &models.Todo{
		ID:          1,
		Title:       "Updated Todo",
		Description: "Updated Description",
		Completed:   true,
	}

	mock.ExpectExec("UPDATE todos SET").
		WithArgs(todo.Title, todo.Description, todo.Completed, sqlmock.AnyArg(), todo.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method being tested
	ctx := context.Background()
	err = repo.Update(ctx, todo)

	// Verify the results
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestDelete(t *testing.T) {
	// Create a new mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database connection: %v", err)
	}
	defer db.Close()

	// Create repository with mock db
	repo := repository.NewPostgresTodoRepository(db)

	// Setup expectations
	mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Call the method being tested
	ctx := context.Background()
	err = repo.Delete(ctx, 1)

	// Verify the results
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestMarkAsCompleted(t *testing.T) {
	// Create a new mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database connection: %v", err)
	}
	defer db.Close()

	// Create repository with mock db
	repo := repository.NewPostgresTodoRepository(db)

	// Setup expectations
	mock.ExpectExec("UPDATE todos SET completed = true, updated_at = \\$1 WHERE id = \\$2").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Call the method being tested
	ctx := context.Background()
	err = repo.MarkAsCompleted(ctx, 1)

	// Verify the results
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// For integration tests, you would need a test database or docker setup
// This is an example of how you would run an integration test with a real database
/*
func TestIntegration_PostgresRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Set up test database connection
	cfg := &config.Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "password",
		DBName:     "todos_test",
	}

	db, err := config.NewPostgresConnection(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Clean up before test
	_, err = db.Exec("DELETE FROM todos")
	if err != nil {
		t.Fatalf("Failed to clean up test database: %v", err)
	}

	// Create repository
	repo := repository.NewPostgresTodoRepository(db)
	ctx := context.Background()

	// Test Create
	todo := &model.Todo{
		Title:       "Integration Test Todo",
		Description: "Testing with real database",
		Completed:   false,
	}

	err = repo.Create(ctx, todo)
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}
	if todo.ID == 0 {
		t.Fatal("Expected todo ID to be set")
	}

	// Test GetByID
	fetchedTodo, err := repo.GetByID(ctx, todo.ID)
	if err != nil {
		t.Fatalf("Failed to get todo: %v", err)
	}
	if fetchedTodo == nil {
		t.Fatal("Expected to get todo, got nil")
	}
	if fetchedTodo.Title != todo.Title {
		t.Errorf("Expected title '%s', got '%s'", todo.Title, fetchedTodo.Title)
	}

*/
