package services

import (
	"github.com/doug-benn/go-server-starter/database"
	"github.com/doug-benn/go-server-starter/models"
)

type Service struct {
	dbController *database.PostgresController
}

func NewService(dbController *database.PostgresController) *Service {
	return &Service{dbController: dbController}
}

func (s *Service) GetAllComments() (*[]models.Comment, error) {
	data, err := s.dbController.GetAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}
