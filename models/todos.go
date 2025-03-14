package models

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

type Todo struct {
	ID          uuid.UUID `json:"id"`
	Description string    `json:"description"`
}
