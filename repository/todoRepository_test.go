package repository_test

// Mock Tests removed as this don't really test much
//below are some example intergration tests that would need a test database to run the tests

/*
func TestIntegration_PostgresRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Set up test database connection
	ctx := context.Background()
	connString := "postgres://postgres:password@localhost:5432/todos_test"

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		t.Fatalf("Unable to parse connection string: %v", err)
	}

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		t.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Create logger
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	// Create PostgresDatabase
	db := &repository.PostgresDatabase{
		Pool:    pool,
		Running: true,
		Logger:  logger,
	}

	// Clean up before test
	_, err = pool.Exec(ctx, "DELETE FROM todos")
	if err != nil {
		t.Fatalf("Failed to clean up test database: %v", err)
	}

	// Create repository
	repo := repository.NewPostgresTodoRepository(db)

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

	// More integration tests...
}
*/
