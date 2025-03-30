package database

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
)

//Postgres Docker Command
//docker run --name postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=test_db -p 5432:5432 -d postgres

// Note: change these and also change the tests
var (
	host     = os.Getenv("POSTGRES_HOST")
	port     = os.Getenv("POSTGRES_PORT")
	username = os.Getenv("POSTGRES_USER")
	password = os.Getenv("POSTGRES_PASSWORD")
	database = os.Getenv("POSTGRES_DB")
)

var pgOnce sync.Once

// Database is the Postgres implementation of the database store.
type PostgresDatabase struct {
	Pool    *pgxpool.Pool
	Running bool
	Logger  zerolog.Logger
}

// NewDatabase creates a database connection pool in DB and pings the database.
func NewDatabase(ctx context.Context, logger zerolog.Logger) (*PostgresDatabase, error) {

	connStr := "postgresql://" + username + ":" + password +
		"@" + host + ":" + port + "/" + database + "?sslmode=disable&connect_timeout=1"

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		logger.Error().AnErr("error", err).Msg("Error parsing pool config")
		return nil, err
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 10 * time.Minute
	config.HealthCheckPeriod = 2 * time.Minute

	var pool *pgxpool.Pool
	pgOnce.Do(func() {
		pool, err = pgxpool.NewWithConfig(ctx, config)
	})

	// Verify the connection
	fails := 0
	const maxFails = 3
	const sleepDuration = 200 * time.Millisecond
	var totalTryTime time.Duration
	for {
		err = pool.Ping(ctx)
		if err == nil {
			break
		} else if ctx.Err() != nil {
			return nil, fmt.Errorf("pinging database: %w", err)
		}
		logger.Error().Err(err).Msg("unable to ping database")
		fails++
		if fails == maxFails {
			return nil, fmt.Errorf("failed connecting to database after %d tries in %s: %w", fails, totalTryTime, err)
		}
		time.Sleep(sleepDuration)
		totalTryTime += sleepDuration
	}

	logger.Info().Msg("Successfully connected to database")
	return &PostgresDatabase{Pool: pool, Running: true, Logger: logger}, nil
}

// Stop stops the database and closes the connection.
func (db *PostgresDatabase) Stop() (err error) {

	db.Logger.Info().Msg("closing database pool")

	if !db.Running {
		db.Logger.Error().Msg("database is not running")
		return fmt.Errorf("%s", "database is not running")
	}

	db.Pool.Close()
	db.Logger.Info().Msg("closing database connection")

	db.Running = false
	return nil
}
