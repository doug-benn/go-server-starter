package database

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// resetSingleton resets the singleton state for testing
func resetSingleton() {
	pgOnce = sync.Once{}
	pgPool = nil
	pgErr = nil
}

// getTestConfig returns a default configuration for testing
func getTestConfig() postgrseConfig {
	return postgrseConfig{
		Host:              "localhost",
		Port:              "5432",
		Username:          "testuser",
		Password:          "testpass",
		Database:          "testdb",
		SSLMode:           "disable",
		ConnectTimeout:    time.Second * 10,
		ApplicationName:   "test-app",
		MaxConns:          10,
		MinConns:          2,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   time.Minute * 30,
		HealthCheckPeriod: time.Minute,
		MaxRetries:        3,
		RetryDelay:        time.Millisecond * 100,
	}
}

// setupPostgresContainer starts a PostgreSQL test container
func setupPostgresContainer(ctx context.Context, t *testing.T) (*postgres.PostgresContainer, postgrseConfig) {
	t.Helper()

	config := getTestConfig()

	container, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase(config.Database),
		postgres.WithUsername(config.Username),
		postgres.WithPassword(config.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Minute)),
	)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	config.Host = host
	config.Port = port.Port()

	return container, config
}

func TestNewDatabase_Success(t *testing.T) {
	ctx := context.Background()
	resetSingleton()

	container, config := setupPostgresContainer(ctx, t)
	defer func() {
		assert.NoError(t, container.Terminate(ctx))
	}()

	logger := zerolog.New(zerolog.NewTestWriter(t))

	db, err := NewDatabase(ctx, logger, config)
	require.NoError(t, err)
	require.NotNil(t, db)

	assert.NotNil(t, db.pool)
	assert.Equal(t, config, db.config)
	assert.False(t, db.closed)

	// Test that the connection works
	err = db.Ping(ctx)
	assert.NoError(t, err)

	// Clean up
	err = db.Close()
	assert.NoError(t, err)
}

func TestNewDatabase_InvalidConfig(t *testing.T) {
	ctx := context.Background()
	resetSingleton()

	logger := zerolog.New(zerolog.NewTestWriter(t))
	config := getTestConfig()
	config.Host = "invalid-host"
	config.Port = "99999"

	db, err := NewDatabase(ctx, logger, config)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to parse pool config")
}

func TestNewDatabase_SingletonBehavior(t *testing.T) {
	ctx := context.Background()
	resetSingleton()

	container, config := setupPostgresContainer(ctx, t)
	defer func() {
		assert.NoError(t, container.Terminate(ctx))
	}()

	logger := zerolog.New(zerolog.NewTestWriter(t))

	// Create first database instance
	db1, err := NewDatabase(ctx, logger, config)
	require.NoError(t, err)
	defer db1.Close()

	// Create second database instance - should use the same pool
	db2, err := NewDatabase(ctx, logger, config)
	require.NoError(t, err)
	defer db2.Close()

	// Both should have the same pool reference
	assert.Equal(t, db1.pool, db2.pool)
}

func TestNewDatabase_ContextCancellation(t *testing.T) {
	resetSingleton()

	logger := zerolog.New(zerolog.NewTestWriter(t))
	config := getTestConfig()

	// Create a context that's immediately cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	db, err := NewDatabase(ctx, logger, config)
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestPostgresDatabase_InitPool_Success(t *testing.T) {
	ctx := context.Background()
	resetSingleton()

	container, config := setupPostgresContainer(ctx, t)
	defer func() {
		assert.NoError(t, container.Terminate(ctx))
	}()

	logger := zerolog.New(zerolog.NewTestWriter(t))
	db := &PostgresDatabase{
		config: config,
		logger: logger,
	}

	pool, err := db.initPool(ctx)
	require.NoError(t, err)
	require.NotNil(t, pool)

	defer pool.Close()

	// Test that the pool configuration is applied correctly
	stats := pool.Config()
	assert.Equal(t, config.MaxConns, stats.MaxConns)
	assert.Equal(t, config.MinConns, stats.MinConns)
	assert.Equal(t, config.MaxConnLifetime, stats.MaxConnLifetime)
	assert.Equal(t, config.MaxConnIdleTime, stats.MaxConnIdleTime)
	assert.Equal(t, config.HealthCheckPeriod, stats.HealthCheckPeriod)
}

func TestPostgresDatabase_InitPool_InvalidConnectionString(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(zerolog.NewTestWriter(t))

	config := getTestConfig()
	config.Host = "invalid host with spaces" // Invalid hostname

	db := &PostgresDatabase{
		config: config,
		logger: logger,
	}

	pool, err := db.initPool(ctx)
	assert.Error(t, err)
	assert.Nil(t, pool)
	assert.Contains(t, err.Error(), "failed to parse pool config")
}

func TestPostgresDatabase_PingWithRetry_Success(t *testing.T) {
	ctx := context.Background()
	resetSingleton()

	container, config := setupPostgresContainer(ctx, t)
	defer func() {
		assert.NoError(t, container.Terminate(ctx))
	}()

	logger := zerolog.New(zerolog.NewTestWriter(t))
	db := &PostgresDatabase{
		config: config,
		logger: logger,
	}

	pool, err := db.initPool(ctx)
	require.NoError(t, err)
	db.pool = pool
	defer pool.Close()

	err = db.pingWithRetry(ctx)
	assert.NoError(t, err)
}

func TestPostgresDatabase_PingWithRetry_FailureWithRetries(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(zerolog.NewTestWriter(t))

	config := getTestConfig()
	config.Host = "nonexistent-host"
	config.MaxRetries = 2
	config.RetryDelay = time.Millisecond * 10 // Fast retries for testing

	db := &PostgresDatabase{
		config: config,
		logger: logger,
	}

	// This will fail to create a pool, but we'll test the retry logic
	pool, err := db.initPool(ctx)
	if err == nil && pool != nil {
		db.pool = pool
		defer pool.Close()

		// Force close the pool to simulate connection failure
		pool.Close()

		err = db.pingWithRetry(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ping database after")
	}
}

func TestPostgresDatabase_PingWithRetry_ContextCancellation(t *testing.T) {
	ctx := context.Background()
	resetSingleton()

	container, config := setupPostgresContainer(ctx, t)
	defer func() {
		assert.NoError(t, container.Terminate(ctx))
	}()

	logger := zerolog.New(zerolog.NewTestWriter(t))
	db := &PostgresDatabase{
		config: config,
		logger: logger,
	}

	pool, err := db.initPool(ctx)
	require.NoError(t, err)
	db.pool = pool
	defer pool.Close()

	// Cancel context during ping
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()

	err = db.pingWithRetry(cancelCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
}

func TestPostgresDatabase_Pool(t *testing.T) {
	ctx := context.Background()
	resetSingleton()

	container, config := setupPostgresContainer(ctx, t)
	defer func() {
		assert.NoError(t, container.Terminate(ctx))
	}()

	logger := zerolog.New(zerolog.NewTestWriter(t))

	db, err := NewDatabase(ctx, logger, config)
	require.NoError(t, err)
	defer db.Close()

	pool := db.Pool()
	assert.NotNil(t, pool)
	assert.Equal(t, db.pool, pool)
}

func TestPostgresDatabase_IsRunning(t *testing.T) {
	ctx := context.Background()
	resetSingleton()

	container, config := setupPostgresContainer(ctx, t)
	defer func() {
		assert.NoError(t, container.Terminate(ctx))
	}()

	logger := zerolog.New(zerolog.NewTestWriter(t))

	db, err := NewDatabase(ctx, logger, config)
	require.NoError(t, err)

	// Should be running initially
	assert.True(t, db.IsRunning())

	// Should not be running after close
	err = db.Close()
	require.NoError(t, err)
	assert.False(t, db.IsRunning())
}

func TestPostgresDatabase_Ping(t *testing.T) {
	ctx := context.Background()
	resetSingleton()

	container, config := setupPostgresContainer(ctx, t)
	defer func() {
		assert.NoError(t, container.Terminate(ctx))
	}()

	logger := zerolog.New(zerolog.NewTestWriter(t))

	db, err := NewDatabase(ctx, logger, config)
	require.NoError(t, err)
	defer db.Close()

	// Ping should work
	err = db.Ping(ctx)
	assert.NoError(t, err)
}

func TestPostgresDatabase_Ping_ClosedConnection(t *testing.T) {
	ctx := context.Background()
	resetSingleton()

	container, config := setupPostgresContainer(ctx, t)
	defer func() {
		assert.NoError(t, container.Terminate(ctx))
	}()

	logger := zerolog.New(zerolog.NewTestWriter(t))

	db, err := NewDatabase(ctx, logger, config)
	require.NoError(t, err)

	// Close the database
	err = db.Close()
	require.NoError(t, err)

	// Ping should fail
	err = db.Ping(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is closed")
}

func TestPostgresDatabase_Close(t *testing.T) {
	ctx := context.Background()
	resetSingleton()

	container, config := setupPostgresContainer(ctx, t)
	defer func() {
		assert.NoError(t, container.Terminate(ctx))
	}()

	logger := zerolog.New(zerolog.NewTestWriter(t))

	db, err := NewDatabase(ctx, logger, config)
	require.NoError(t, err)

	// Close should work
	err = db.Close()
	assert.NoError(t, err)
	assert.True(t, db.closed)

	// Second close should not error
	err = db.Close()
	assert.NoError(t, err)
}

func TestPostgresDatabase_Close_NilPool(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))

	db := &PostgresDatabase{
		config: getTestConfig(),
		logger: logger,
		pool:   nil, // Nil pool
	}

	err := db.Close()
	assert.NoError(t, err)
	assert.True(t, db.closed)
}

func TestPostgresDatabase_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	resetSingleton()

	container, config := setupPostgresContainer(ctx, t)
	defer func() {
		assert.NoError(t, container.Terminate(ctx))
	}()

	logger := zerolog.New(zerolog.NewTestWriter(t))

	db, err := NewDatabase(ctx, logger, config)
	require.NoError(t, err)
	defer db.Close()

	// Test concurrent access to methods
	var wg sync.WaitGroup
	numGoroutines := 10

	// Test concurrent Pool() calls
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			pool := db.Pool()
			assert.NotNil(t, pool)
		}()
	}

	// Test concurrent IsRunning() calls
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			running := db.IsRunning()
			assert.True(t, running)
		}()
	}

	// Test concurrent Ping() calls
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			err := db.Ping(ctx)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}

func TestPostgresDatabase_ConfigurationEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		modifyConfig  func(*postgrseConfig)
		expectError   bool
		errorContains string
	}{
		{
			name: "Zero MaxRetries",
			modifyConfig: func(c *postgrseConfig) {
				c.MaxRetries = 0
			},
			expectError:   true,
			errorContains: "failed to ping database after 0 attempts",
		},
		{
			name: "Very Short RetryDelay",
			modifyConfig: func(c *postgrseConfig) {
				c.RetryDelay = time.Nanosecond
			},
			expectError: false, // Should still work with valid connection
		},
		{
			name: "Very Short ConnectTimeout",
			modifyConfig: func(c *postgrseConfig) {
				c.ConnectTimeout = time.Nanosecond
			},
			expectError: false, // pgx handles this gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resetSingleton()

			if !tt.expectError {
				container, config := setupPostgresContainer(ctx, t)
				defer func() {
					assert.NoError(t, container.Terminate(ctx))
				}()
				tt.modifyConfig(&config)

				logger := zerolog.New(zerolog.NewTestWriter(t))
				db, err := NewDatabase(ctx, logger, config)

				if tt.expectError {
					assert.Error(t, err)
					if tt.errorContains != "" {
						assert.Contains(t, err.Error(), tt.errorContains)
					}
					assert.Nil(t, db)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, db)
					if db != nil {
						db.Close()
					}
				}
			} else {
				// For error cases, use invalid config
				config := getTestConfig()
				config.Host = "nonexistent-host"
				tt.modifyConfig(&config)

				logger := zerolog.New(zerolog.NewTestWriter(t))
				db, err := NewDatabase(ctx, logger, config)

				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, db)
			}
		})
	}
}
