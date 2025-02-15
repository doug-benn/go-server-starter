package database

import (
	"database/sql"

	"github.com/doug-benn/go-server-starter/entity"
	"github.com/google/uuid"
)

type PostgresTaskRepo struct {
	db *sql.DB
}

func NewPostgresTaskRepo(db *sql.DB) *PostgresTaskRepo {
	return &PostgresTaskRepo{db: db}
}

func (p *PostgresTaskRepo) GetByID(id uuid.UUID) (entity.Task, error) {
	return entity.Task{}, nil
}
