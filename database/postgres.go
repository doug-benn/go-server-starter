package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

//Postgres Docker Command
//docker run --name postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=test_db -p 5432:5432 -d postgres

type postgreConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func (p *postgreConfig) loadPostgresConfig() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	p.Host = os.Getenv("POSTGRES_HOST")
	p.Port = os.Getenv("POSTGRES_PORT")
	p.User = os.Getenv("POSTGRES_USER")
	p.Password = os.Getenv("POSTGRES_PASSWORD")
	p.Database = os.Getenv("POSTGRES_DB")
}

// Database is the Postgres implementation of the database store.
type Database struct {
	startStopMutex sync.Mutex
	running        bool
	sql            *sql.DB
	logger         *slog.Logger
}

// NewDatabase creates a database connection pool in DB and pings the database.
func NewDatabase(log *slog.Logger, connLimits bool, idleLimits bool) (*Database, error) {
	config := postgreConfig{}
	config.loadPostgresConfig()

	connStr := "postgresql://" + config.User + ":" + config.Password +
		"@" + config.Host + ":" + config.Port + "/" + config.Database + "?sslmode=disable&connect_timeout=1"

	fmt.Println(connStr)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if connLimits {
		db.SetMaxOpenConns(5)
	}

	if idleLimits {
		db.SetMaxIdleConns(5)
	}

	return &Database{
		sql:    db,
		logger: log,
	}, nil
}

// Start pings the database, and if it fails, retries up to 3 times
// before returning a start error.
func (db *Database) Start(ctx context.Context) (runError <-chan error, err error) {
	db.startStopMutex.Lock()
	defer db.startStopMutex.Unlock()

	if db.running {
		return nil, fmt.Errorf("%s", "database service is already running")
	}

	fails := 0
	const maxFails = 3
	const sleepDuration = 200 * time.Millisecond
	var totalTryTime time.Duration
	for {
		err = db.sql.PingContext(ctx)
		if err == nil {
			break
		} else if ctx.Err() != nil {
			return nil, fmt.Errorf("pinging database: %w", err)
		}
		fails++
		if fails == maxFails {
			return nil, fmt.Errorf("failed connecting to database after %d tries in %s: %w", fails, totalTryTime, err)
		}
		time.Sleep(sleepDuration)
		totalTryTime += sleepDuration
	}

	db.running = true

	// TODO have periodic ping to check connection is still alive and signal through the run error channel.
	return nil, nil
}

// Stop stops the database and closes the connection.
func (db *Database) Stop() (err error) {
	db.startStopMutex.Lock()
	defer db.startStopMutex.Unlock()
	if !db.running {
		return fmt.Errorf("%s", "database is not running")
	}

	err = db.sql.Close()
	if err != nil {
		return fmt.Errorf("closing database connection: %w", err)
	}

	db.running = false
	return nil
}
