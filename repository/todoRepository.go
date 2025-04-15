package repository

import (
	"context"

	"github.com/doug-benn/go-server-starter/models"
)

type TodoRepository interface {
	Create(ctx context.Context, todo *models.Todo) error
	GetByID(ctx context.Context, id int64) (*models.Todo, error)
	GetAll(ctx context.Context) (models.Todos, error)
	Update(ctx context.Context, todo *models.Todo) error
	Delete(ctx context.Context, id int64) error
	MarkAsCompleted(ctx context.Context, id int64) error
}
