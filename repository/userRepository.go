package repository

import (
	"fmt"

	"github.com/doug-benn/go-server-starter/database"
	"github.com/doug-benn/go-server-starter/models"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Get(id int) (*models.User, error)
	Save(user *models.User) error
}

// InMemoryUserRepository is an in-memory implementation of UserRepository.
type InMemoryUserRepository struct {
	users map[int]*models.User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{users: make(map[int]*models.User)}
}

func (r *InMemoryUserRepository) Get(id int) (*models.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (r *InMemoryUserRepository) Save(user *models.User) error {
	r.users[user.ID] = user
	return nil
}

// DatabaseUserRepository is a database implementation of UserRepository.
type DatabaseUserRepository struct {
	db *database.PostgresDatabase
}

func NewDatabaseUserRepository(db *database.PostgresDatabase) *DatabaseUserRepository {
	return &DatabaseUserRepository{db: db}
}

func (r *DatabaseUserRepository) Get(id int) (*models.User, error) {
	if id == 1 {
		return &models.User{ID: 1, Name: "Database User"}, nil
	}
	return nil, fmt.Errorf("user not found in database")
}

func (r *DatabaseUserRepository) Save(user *models.User) error {
	fmt.Println("Saving user to database:", user)
	return nil
}
