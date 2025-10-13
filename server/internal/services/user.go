package services

import (
	"context"
	"fmt"
	"time"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/database"
	"github.com/datazip/olake-ui/server/internal/models"
)

type UserService struct {
	userORM *database.UserORM
}

func NewUserService() *UserService {
	return &UserService{
		userORM: database.NewUserORM(),
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *models.User) error {
	if err := s.userORM.Create(req); err != nil {
		return fmt.Errorf("failed to create user: %s", err)
	}
	return nil
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	users, err := s.userORM.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve users: %s", err)
	}
	return users, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id int, req *models.User) (*models.User, error) {
	existingUser, err := s.userORM.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	existingUser.Username = req.Username
	existingUser.Email = req.Email
	existingUser.UpdatedAt = time.Now()

	if err := s.userORM.Update(existingUser); err != nil {
		return nil, fmt.Errorf("%s user: %s", constants.ErrFailedToUpdate, err)
	}

	return existingUser, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	if err := s.userORM.Delete(id); err != nil {
		return fmt.Errorf("failed to delete user: %s", err)
	}
	return nil
}

func (s *UserService) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	user, err := s.userORM.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}
