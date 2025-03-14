package services

import (
	"database/sql"

	"github.com/doug-benn/go-server-starter/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type TodoRepository struct {
	DB *sql.DB
}

func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{DB: db}
}

func (r *TodoRepository) GetAll() ([]*models.Todo, error) {
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

func (r *TodoRepository) GetByID(id uuid.UUID) (*models.Todo, error) {
	t := new(models.Todo)
	err := r.DB.QueryRow("SELECT id, description FROM todos WHERE id = $1", id).Scan(&t.ID, &t.Description)
	if err != nil {
		return nil, err
	}
	return t, nil
}
