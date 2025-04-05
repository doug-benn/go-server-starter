package services

import (
	"testing"

	"github.com/doug-benn/go-server-starter/models"
	"github.com/doug-benn/go-server-starter/repository"
)

func TestUserService_GetUser_InMemory(t *testing.T) {
	repo := repository.NewInMemoryUserRepository()
	service := NewUserService(repo)

	user := &models.User{ID: 1, Name: "Test User"}
	err := service.CreateUser(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrievedUser, err := service.GetUser(1)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrievedUser.ID != user.ID || retrievedUser.Name != user.Name {
		t.Errorf("Retrieved user does not match expected user. Expected: %+v, Got: %+v", user, retrievedUser)
	}
}

func TestUserService_GetUser_NotFound(t *testing.T) {
	repo := repository.NewInMemoryUserRepository()
	service := NewUserService(repo)

	_, err := service.GetUser(999) // Non-existent ID
	if err == nil {
		t.Errorf("Expected error for non-existent user, got nil")
	}

	if err.Error() != "user not found" {
		t.Errorf("Expected error message 'user not found', got '%s'", err.Error())
	}
}

func TestUserService_CreateUser(t *testing.T) {
	repo := repository.NewInMemoryUserRepository()
	service := NewUserService(repo)

	user := &models.User{ID: 1, Name: "Test User"}
	err := service.CreateUser(user)

	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrievedUser, err := service.GetUser(1)

	if retrievedUser.ID != user.ID || retrievedUser.Name != user.Name {
		t.Errorf("Retrieved user does not match expected user. Expected: %+v, Got: %+v", user, retrievedUser)
	}

}

// func TestUserService_DatabaseGetUser(t *testing.T) {
// 	repo := repository.NewDatabaseUserRepository()
// 	service := NewUserService(repo)

// 	user, err := service.GetUser(1)

// 	if err != nil {
// 		t.Fatalf("Failed to get user: %v", err)
// 	}

// 	if user.ID != 1 || user.Name != "Database User" {
// 		t.Errorf("Retrieved user does not match expected user. Expected: %+v, Got: %+v", models.User{ID: 1, Name: "Database User"}, user)
// 	}

// 	_, err = service.GetUser(2)

// 	if err == nil {
// 		t.Fatalf("Expected error for non existing user, got nil")
// 	}
// }
