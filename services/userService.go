package services

import (
	"github.com/doug-benn/go-server-starter/models"
	"github.com/doug-benn/go-server-starter/repository"
)

// UserService uses the UserRepository interface.
type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(id int) (*models.User, error) {
	return s.repo.Get(id)
}

func (s *UserService) CreateUser(user *models.User) error {
	return s.repo.Save(user)
}
