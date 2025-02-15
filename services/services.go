package services

import (
	"github.com/doug-benn/go-server-starter/entity"
	"github.com/google/uuid"
)

type TaskService struct {
	repo entity.TaskRepo
}

func NewTaskService(r entity.TaskRepo) *TaskService {
	return &TaskService{repo: r}
}

func (s *TaskService) GetByID(id uuid.UUID) entity.Task {
	return entity.Task{}
}
