package services

import (
	"github.com/doug-benn/go-server-starter/models"
	"github.com/doug-benn/go-server-starter/repository"
)

// UserService uses the UserRepository interface.
type TodoService struct {
	repo repository.TodoRepository
}

func NewTodoService(repo repository.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

func (s *TodoService) GetAllTodos() (models.Todos, error) {
	return s.repo.GetTodos()
}

func (s *TodoService) GetById(id int) (*models.Todo, error) {
	return s.repo.GetTodoById(id)
}
