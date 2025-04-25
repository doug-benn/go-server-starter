package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/doug-benn/go-server-starter/database"
	"github.com/doug-benn/go-server-starter/models"
	_ "github.com/lib/pq"
)

type PostgresTodoRepository struct {
	db *database.PostgresDatabase
}

func NewPostgresTodoRepository(db *database.PostgresDatabase) *PostgresTodoRepository {
	return &PostgresTodoRepository{
		db: db,
	}
}

func (r *PostgresTodoRepository) Create(ctx context.Context, todo *models.Todo) error {
	query := `
		INSERT INTO todos (title, description, completed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	now := time.Now()
	todo.CreatedAt = now
	todo.UpdatedAt = now

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		todo.Title,
		todo.Description,
		todo.Completed,
		todo.CreatedAt,
		todo.UpdatedAt,
	).Scan(&todo.ID)

	return err
}

func (r *PostgresTodoRepository) GetByID(ctx context.Context, id int64) (*models.Todo, error) {
	query := `
		SELECT id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE id = $1
	`

	todo := &models.Todo{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return todo, nil
}

func (r *PostgresTodoRepository) GetAll(ctx context.Context) (models.Todos, error) {
	query := `
		SELECT id, title, description, completed, created_at, updated_at
		FROM todos
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos models.Todos
	for rows.Next() {
		todo := models.Todo{}
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func (r *PostgresTodoRepository) Update(ctx context.Context, todo *models.Todo) error {
	query := `
		UPDATE todos
		SET title = $1, description = $2, completed = $3, updated_at = $4
		WHERE id = $5
	`

	todo.UpdatedAt = time.Now()

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		todo.Title,
		todo.Description,
		todo.Completed,
		todo.UpdatedAt,
		todo.ID,
	)

	return err
}

func (r *PostgresTodoRepository) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE FROM todos
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *PostgresTodoRepository) MarkAsCompleted(ctx context.Context, id int64) error {
	query := `
		UPDATE todos
		SET completed = true, updated_at = $1
		WHERE id = $2
	`

	now := time.Now()
	_, err := r.db.Pool.Exec(ctx, query, now, id)
	return err
}
