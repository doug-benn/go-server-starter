package services

import (
	"context"
	"fmt"
	"time"

	"github.com/doug-benn/go-server-starter/database"
	"github.com/doug-benn/go-server-starter/models"
)

type todoRepository struct {
	db *database.PostgresDatabase
}

func NewTodoRepository(db *database.PostgresDatabase) models.TodoService {
	return &todoRepository{db: db}
}

func (r *todoRepository) GetTodos() (models.Todos, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := r.db.Pool.Query(ctx, "SELECT id, task FROM todos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos models.Todos
	for rows.Next() {
		var todo models.Todo
		err = rows.Scan(&todo.ID, &todo.Task)
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

func (r *todoRepository) GetTodoById(id int) (models.Todo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	row := r.db.Pool.QueryRow(ctx, "SELECT id, task FROM todos WHERE id = $1", id)
	if row == nil {
		return models.Todo{}, fmt.Errorf("todo with id %d not found", id)
	}
	var todo models.Todo
	err := row.Scan(&todo.ID, &todo.Task)
	if err != nil {
		return models.Todo{}, err
	}

	return todo, nil
}

func (r *todoRepository) CreateTodo(todo models.Todo) (models.Todo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := r.db.Pool.QueryRow(ctx, "INSERT INTO todos (task) VALUES ($1) RETURNING id", todo.Task).Scan(&todo.ID)
	if err != nil {
		return models.Todo{}, err
	}

	return todo, nil
}

func (r *todoRepository) UpdateTodo(todo models.Todo) (models.Todo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.db.Pool.Exec(ctx, "UPDATE todos SET task=$1 WHERE id=$2", todo.Task, todo.ID)
	if err != nil {
		return models.Todo{}, err
	}

	return todo, nil
}

func (r *todoRepository) DeleteTodo(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.db.Pool.Exec(ctx, "DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return err
	}

	return nil
}
