package services

import (
	"fmt"
	"time"

	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/models"
)

type UserService struct {
	userORM *database.UserORM
}

func NewUserService() *UserService {
	return &UserService{
		userORM: database.NewUserORM(),
	}
}

func (s *UserService) CreateUser(user *models.User) error {
	return s.userORM.Create(user)
}

func (s *UserService) GetAllUsers() ([]*models.User, error) {
	return s.userORM.GetAll()
}

func (s *UserService) UpdateUser(id int, updateReq *models.User) (*models.User, error) {
	// Get existing user
	existingUser, err := s.userORM.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrUserNotFound, err)
	}

	// Update fields
	existingUser.Username = updateReq.Username
	existingUser.Email = updateReq.Email
	existingUser.UpdatedAt = time.Now()

	if err := s.userORM.Update(existingUser); err != nil {
		return nil, fmt.Errorf("%s user: %w", ErrFailedToUpdate, err)
	}

	return existingUser, nil
}

func (s *UserService) DeleteUser(id int) error {
	return s.userORM.Delete(id)
}

func (s *UserService) GetUserByID(id int) (*models.User, error) {
	return s.userORM.GetByID(id)
}
