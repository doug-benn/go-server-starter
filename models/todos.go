package models

import (
	"errors"
	"time"
)

var (
	ErrInvalidTask     = errors.New("task must have a todo text")
	ErrTaskNotFound    = errors.New("the task was not found in the repository")
	ErrFailedToAddTask = errors.New("failed to add the task to the repository")
	ErrUpdatingTask    = errors.New("failed to update the task in the repository")
)

type Todo struct {
	ID        int64     `json:"id"`
	Todo      string    `json:"todo"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Helper methods for status checking
func (t *Todo) IsCompleted() bool {
	return t.Status == "completed"
}

func (t *Todo) IsPending() bool {
	return t.Status == "pending"
}

func (t *Todo) IsInProgress() bool {
	return t.Status == "in_progress"
}

type Todos []Todo
