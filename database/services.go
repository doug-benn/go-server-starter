package database

import (
	"context"

	"github.com/doug-benn/go-server-starter/models"
)

// Service represents a service that interacts with a database.
type PostgresService interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	Start(context.Context) (<-chan error, error)

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Stop() error

	GetAllComments() ([]models.Comment, error)
}
