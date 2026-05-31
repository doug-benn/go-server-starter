package e2e

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/doug-benn/go-server-starter/database"
	"github.com/doug-benn/go-server-starter/producer"
	"github.com/doug-benn/go-server-starter/repository"
	"github.com/doug-benn/go-server-starter/sse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func loadMigrationSQL() (string, error) {
	b, err := os.ReadFile("../migrations/000001_create_todo_table.up.sql")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func setupPostgresContainer(t *testing.T) (database.PostgresConfig, func()) {
	t.Helper()

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	config := database.PostgresConfig{
		Host:              host,
		Port:              int(port.Num()),
		Username:          "testuser",
		Password:          "testpass",
		Database:          "testdb",
		SSLMode:           "disable",
		ConnectTimeout:    5 * time.Second,
		ApplicationName:   "test-sse-e2e",
		MaxConns:          10,
		MinConns:          1,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
		MaxRetries:        1,
		InitialRetryDelay: 100 * time.Millisecond,
	}

	cleanup := func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return config, cleanup
}

func setupSSEPipeline(t *testing.T, ctx context.Context) (*database.PostgresDatabase, *producer.Producer[sse.Event], func()) {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	config, cleanupContainer := setupPostgresContainer(t)

	db, err := database.NewDatabase(ctx, logger, config)
	require.NoError(t, err)

	migrationSQL, err := loadMigrationSQL()
	require.NoError(t, err)
	_, err = db.Pool().Exec(ctx, migrationSQL)
	require.NoError(t, err)

	sseProducer := producer.NewProducer(
		producer.WithBroadcastTimeout[sse.Event](time.Second),
		producer.WithCustomLogger[sse.Event](logger),
	)
	go sseProducer.Start(ctx)

	postgresListener := database.NewListener(db.Pool())
	err = postgresListener.Connect(ctx)
	require.NoError(t, err)

	err = postgresListener.ListenToChannel(ctx, "events")
	require.NoError(t, err)

	notifyCtx, notifyCancel := context.WithCancel(ctx)
	go repository.NotificationProcessing(notifyCtx, logger, postgresListener, sseProducer)

	cleanup := func() {
		notifyCancel()
		time.Sleep(50 * time.Millisecond)
		postgresListener.Close(ctx)
		db.Close()
		cleanupContainer()
	}

	return db, sseProducer, cleanup
}

func consumeInsertEvent(ctx context.Context, t *testing.T, sub *producer.Subscription[sse.Event], title, description string) {
	t.Helper()
	eventCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	event, err := sub.Next(eventCtx)
	require.NoError(t, err)

	dbEvent, ok := event.Data.(*repository.DatabaseEvent)
	require.True(t, ok)
	assert.Equal(t, "todos", dbEvent.Table)
	assert.Equal(t, "INSERT", dbEvent.Action)
	assert.Equal(t, title, dbEvent.Data["title"])
	assert.Equal(t, description, dbEvent.Data["description"])
}

func TestSSETodoCreated(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	db, sseProducer, cleanup := setupSSEPipeline(t, ctx)
	defer cleanup()

	sub := sseProducer.Subscribe(100)

	repo := repository.New(db.Pool())
	todo, err := repo.CreateTodo(ctx, repository.CreateTodoParams{
		Title:       "E2E Test Todo",
		Description: "Testing the SSE pipeline end to end",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	require.NoError(t, err)
	require.NotZero(t, todo.ID)

	eventCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	event, err := sub.Next(eventCtx)
	require.NoError(t, err)

	dbEvent, ok := event.Data.(*repository.DatabaseEvent)
	require.True(t, ok, "event data should be *DatabaseEvent")
	assert.Equal(t, "todos", dbEvent.Table)
	assert.Equal(t, "INSERT", dbEvent.Action)
	assert.False(t, dbEvent.Timestamp.IsZero())
	assert.Equal(t, todo.Title, dbEvent.Data["title"])
	assert.Equal(t, todo.Description, dbEvent.Data["description"])
	assert.Equal(t, false, dbEvent.Data["completed"])
}

func TestSSETodoUpdated(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	db, sseProducer, cleanup := setupSSEPipeline(t, ctx)
	defer cleanup()

	sub := sseProducer.Subscribe(100)

	repo := repository.New(db.Pool())

	todo, err := repo.CreateTodo(ctx, repository.CreateTodoParams{
		Title:       "Update Test",
		Description: "Will be completed",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	require.NoError(t, err)

	consumeInsertEvent(ctx, t, sub, "Update Test", "Will be completed")

	_, err = repo.CompleteTodo(ctx, repository.CompleteTodoParams{
		UpdatedAt: time.Now(),
		ID:        todo.ID,
	})
	require.NoError(t, err)

	eventCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	event, err := sub.Next(eventCtx)
	require.NoError(t, err)

	dbEvent, ok := event.Data.(*repository.DatabaseEvent)
	require.True(t, ok)
	assert.Equal(t, "todos", dbEvent.Table)
	assert.Equal(t, "UPDATE", dbEvent.Action)
	assert.Equal(t, float64(todo.ID), dbEvent.Data["id"])
	assert.Equal(t, true, dbEvent.Data["completed"])
}

func TestSSETodoDeleted(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	db, sseProducer, cleanup := setupSSEPipeline(t, ctx)
	defer cleanup()

	sub := sseProducer.Subscribe(100)

	repo := repository.New(db.Pool())

	todo, err := repo.CreateTodo(ctx, repository.CreateTodoParams{
		Title:       "Delete Test",
		Description: "Will be deleted",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	require.NoError(t, err)

	consumeInsertEvent(ctx, t, sub, "Delete Test", "Will be deleted")

	err = repo.DeleteTodo(ctx, todo.ID)
	require.NoError(t, err)

	eventCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	event, err := sub.Next(eventCtx)
	require.NoError(t, err)

	dbEvent, ok := event.Data.(*repository.DatabaseEvent)
	require.True(t, ok)
	assert.Equal(t, "todos", dbEvent.Table)
	assert.Equal(t, "DELETE", dbEvent.Action)
	assert.Equal(t, float64(todo.ID), dbEvent.Data["id"])
}

func TestMultipleSSEClients(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	db, sseProducer, cleanup := setupSSEPipeline(t, ctx)
	defer cleanup()

	subA := sseProducer.Subscribe(100)
	subB := sseProducer.Subscribe(100)

	repo := repository.New(db.Pool())

	_, err := repo.CreateTodo(ctx, repository.CreateTodoParams{
		Title:       "Multi-Client Test",
		Description: "Testing broadcast to multiple subscribers",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	require.NoError(t, err)

	consumeInsertEvent(ctx, t, subA, "Multi-Client Test", "Testing broadcast to multiple subscribers")
	consumeInsertEvent(ctx, t, subB, "Multi-Client Test", "Testing broadcast to multiple subscribers")
}
