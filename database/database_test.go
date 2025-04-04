package database

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func mustStartPostgresContainer() (func(context.Context, ...testcontainers.TerminateOption) error, error) {
	var (
		dbName = "database"
		dbPwd  = "password"
		dbUser = "user"
	)

	dbContainer, err := postgres.Run(
		context.Background(),
		"postgres:latest",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPwd),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	database = dbName
	password = dbPwd
	username = dbUser

	dbHost, err := dbContainer.Host(context.Background())
	if err != nil {
		return dbContainer.Terminate, err
	}

	dbPort, err := dbContainer.MappedPort(context.Background(), "5432/tcp")
	if err != nil {
		return dbContainer.Terminate, err
	}

	host = dbHost
	port = dbPort.Port()

	return dbContainer.Terminate, err
}

func TestMain(m *testing.M) {
	teardown, err := mustStartPostgresContainer()
	if err != nil {
		log.Fatalf("could not start postgres container: %v", err)
	}

	m.Run()

	if teardown != nil && teardown(context.Background()) != nil {
		log.Fatalf("could not teardown postgres container: %v", err)
	}
}

func TestNew(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	srv, err := NewDatabase(ctx, zerolog.New(os.Stdout))
	if err != nil {
		t.Fatal("NewDatabase() returned an err", err)
	}
	if srv == nil {
		t.Fatal("NewDatabase() returned nil")
	}
}

func TestClose(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	srv, err := NewDatabase(ctx, zerolog.New(os.Stdout))
	if err != nil {
		t.Fatal("NewDatabase() returned an err", err)
	}

	if srv.Stop() != nil {
		t.Fatalf("expected Stop() to return nil")
	}
}

// func TestHealth(t *testing.T) {
// 	srv, err := NewDatabase(true, true)
// 	if err != nil {
// 		t.Fatal("NewDatabase() returned an err", err)
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	srv.Start(ctx)

// 	stats := srv.Health()

// 	if stats["status"] != "up" {
// 		t.Fatalf("expected status to be up, got %s", stats["status"])
// 	}

// 	if _, ok := stats["error"]; ok {
// 		t.Fatalf("expected error not to be present")
// 	}

// 	if stats["message"] != "It's healthy" {
// 		t.Fatalf("expected message to be 'It's healthy', got %s", stats["message"])
// 	}
// }
