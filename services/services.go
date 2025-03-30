package services

import (
	"database/sql"

	"github.com/doug-benn/go-server-starter/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type todoRepo struct {
	DB *sql.DB
}

func NewDataService(db *sql.DB) DataService {
	return &todoRepo{DB: db}
}

type DataService interface {
	GetAll() ([]*models.Todo, error)
	GetByID(id uuid.UUID) (*models.Todo, error)
}

func (r *todoRepo) GetAll() ([]*models.Todo, error) {
	var todos []*models.Todo
	rows, err := r.DB.Query("SELECT id, description FROM todos")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		t := new(models.Todo)
		err = rows.Scan(&t.ID, &t.Description)
		if err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, nil
}

func (r *todoRepo) GetByID(id uuid.UUID) (*models.Todo, error) {
	t := new(models.Todo)
	err := r.DB.QueryRow("SELECT id, description FROM todos WHERE id = $1", id).Scan(&t.ID, &t.Description)
	if err != nil {
		return nil, err
	}
	return t, nil
}
