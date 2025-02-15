package entity

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidTask     = errors.New("task must have a title and description")
	ErrTaskNotFound    = errors.New("the task was not found in the repository")
	ErrFailedToAddTask = errors.New("failed to add the task to the repository")
	ErrUpdatingTask    = errors.New("failed to update the task in the repository")
)

type Task struct {
	id              uuid.UUID
	taskTitle       string
	taskDescription string
	status          string
}

type TaskRepo interface {
	GetByID(uuid.UUID) (Task, error)
}
