package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/doug-benn/go-server-starter/utilities"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
)

var (
	pgOnce sync.Once
	pgPool *pgxpool.Pool
	pgErr  error
)

// Config holds database configuration parameters
type postgrseConfig struct {
	Host              string
	Port              string
	Username          string
	Password          string
	Database          string
	SSLMode           string
	ConnectTimeout    time.Duration
	ApplicationName   string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	MaxRetries        int
	RetryDelay        time.Duration
}

// NewConfig returns a configuration with sensible defaults for a locally host postgres database
func NewConfig() postgrseConfig {
	return postgrseConfig{
		Host:              utilities.GetEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:              utilities.GetEnvOrDefault("POSTGRES_PORT", "5432"),
		Username:          utilities.GetEnvOrDefault("POSTGRES_USER", "postgres"),
		Password:          utilities.GetEnvOrDefault("POSTGRES_PASSWORD", "password"),
		Database:          utilities.GetEnvOrDefault("POSTGRES_DB", "testdb"),
		SSLMode:           utilities.GetEnvOrDefault("POSTGRES_SSL_MODE", "disable"),
		ConnectTimeout:    5 * time.Second,
		ApplicationName:   utilities.GetEnvOrDefault("APPLICATION_NAME", "go-server-starter"),
		MaxConns:          30,
		MinConns:          1,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
		MaxRetries:        3,
		RetryDelay:        2 * time.Second,
	}
}

// PostgresDatabase is the Postgres implementation of the database store.
type PostgresDatabase struct {
	pool   *pgxpool.Pool
	config postgrseConfig
	logger zerolog.Logger
	mu     sync.RWMutex
	closed bool
}

// NewDatabase creates a database connection pool and pings the database.
func NewDatabase(ctx context.Context, logger zerolog.Logger, config postgrseConfig) (*PostgresDatabase, error) {

	db := &PostgresDatabase{
		config: config,
		logger: logger,
	}

	// Initialize the connection pool using singleton pattern
	pgOnce.Do(func() {
		pgPool, pgErr = db.initPool(ctx)
	})

	if pgErr != nil {
		return nil, fmt.Errorf("failed to initialize database pool: %w", pgErr)
	}

	if pgPool == nil {
		return nil, fmt.Errorf("database pool is nil after initialization")
	}

	db.pool = pgPool

	// Test the connection with retry logic
	if err := db.pingWithRetry(ctx); err != nil {
		return nil, fmt.Errorf("failed to establish database connection: %w", err)
	}

	logger.Info().Msg("successfully connected to database")
	return db, nil
}

// initPool creates and configures the connection pool
func (db *PostgresDatabase) initPool(ctx context.Context) (*pgxpool.Pool, error) {

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s&connect_timeout=%d&application_name=%s",
		db.config.Username,
		db.config.Password,
		db.config.Host,
		db.config.Port,
		db.config.Database,
		db.config.SSLMode,
		int(db.config.ConnectTimeout.Seconds()),
		db.config.ApplicationName,
	)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		db.logger.Error().Err(err).Msg("error parsing pool config")
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	// Configure connection pool settings
	config.MaxConns = db.config.MaxConns
	config.MinConns = db.config.MinConns
	config.MaxConnLifetime = db.config.MaxConnLifetime
	config.MaxConnIdleTime = db.config.MaxConnIdleTime
	config.HealthCheckPeriod = db.config.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return pool, nil
}

// pingWithRetry attempts to ping the database with exponential backoff
func (db *PostgresDatabase) pingWithRetry(ctx context.Context) error {
	var lastErr error
	delay := db.config.RetryDelay

	for attempt := 1; attempt <= db.config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during database ping: %w", ctx.Err())
		default:
		}

		if err := db.pool.Ping(ctx); err == nil {
			return nil // Success
		} else {
			lastErr = err
			db.logger.Warn().
				Err(err).
				Int("attempt", attempt).
				Int("max_attempts", db.config.MaxRetries).
				Dur("retry_delay", delay).
				Msg("database ping failed, retrying")

			if attempt < db.config.MaxRetries {
				select {
				case <-ctx.Done():
					return fmt.Errorf("context cancelled during retry delay: %w", ctx.Err())
				case <-time.After(delay):
					delay *= 2 // Exponential backoff
				}
			}
		}
	}

	return fmt.Errorf("failed to ping database after %d attempts: %w", db.config.MaxRetries, lastErr)
}

// Pool returns the underlying connection pool (read-only access)
func (db *PostgresDatabase) Pool() *pgxpool.Pool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.pool
}

// IsRunning returns true if the database connection is active
func (db *PostgresDatabase) IsRunning() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return !db.closed && db.pool != nil
}

// Ping tests the database connection
func (db *PostgresDatabase) Ping(ctx context.Context) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return fmt.Errorf("database connection is closed")
	}

	return db.pool.Ping(ctx)
}

// Close closes the connection pool
func (db *PostgresDatabase) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		db.logger.Warn().Msg("database connection already closed")
		return nil
	}

	db.logger.Info().Msg("closing database connection pool")

	if db.pool != nil {
		db.pool.Close()
	}

	db.closed = true
	db.logger.Info().Msg("database connection pool closed successfully")

	return nil
}
