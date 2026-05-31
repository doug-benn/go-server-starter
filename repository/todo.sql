-- name: GetTodo :one
SELECT id, title, description, completed, created_at, updated_at
FROM todos
WHERE id = $1;

-- name: ListTodos :many
SELECT id, title, description, completed, created_at, updated_at
FROM todos
ORDER BY created_at DESC;

-- name: CreateTodo :one
INSERT INTO todos (title, description, completed, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, title, description, completed, created_at, updated_at;

-- name: UpdateTodo :one
UPDATE todos
SET title = $1, description = $2, completed = $3, updated_at = $4
WHERE id = $5
RETURNING id, title, description, completed, created_at, updated_at;

-- name: DeleteTodo :exec
DELETE FROM todos
WHERE id = $1;

-- name: CompleteTodo :one
UPDATE todos
SET completed = true, updated_at = $1
WHERE id = $2
RETURNING id, title, description, completed, created_at, updated_at;
