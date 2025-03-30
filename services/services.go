package services

import (
	"database/sql"
	"fmt"

	"github.com/doug-benn/go-server-starter/models"
	_ "github.com/lib/pq"
)

type todoRepository struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) models.TodoService {
	return &todoRepository{db: db}
}

func (r *todoRepository) GetTodos() (models.Todos, error) {
	rows, err := r.db.Query("SELECT id, task FROM todos")
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
	row := r.db.QueryRow("SELECT id, task FROM todos WHERE id = $1", id)
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
	result, err := r.db.Exec("INSERT INTO todos (task) VALUES ($1)", todo.Task)
	if err != nil {
		return models.Todo{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return models.Todo{}, err
	}

	todo.ID = int(id)

	return todo, nil
}

func (r *todoRepository) UpdateTodo(todo models.Todo) (models.Todo, error) {
	_, err := r.db.Exec("UPDATE todos SET task=$1 WHERE id=$2", todo.Task, todo.ID)
	if err != nil {
		return models.Todo{}, err
	}

	return todo, nil
}

func (r *todoRepository) DeleteTodo(id int) error {
	_, err := r.db.Exec("DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return err
	}

	return nil
}
