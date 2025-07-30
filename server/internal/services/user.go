package services

import (
	"context"
	"fmt"
	"time"

	"github.com/beego/beego/v2/core/logs"
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

func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	if err := s.userORM.Create(user); err != nil {
		logs.Error("Failed to create user: %v", err)
		return fmt.Errorf("failed to create user: %s", err)
	}
	return nil
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	users, err := s.userORM.GetAll()
	if err != nil {
		logs.Error("Failed to retrieve users: %v", err)
		return nil, fmt.Errorf("failed to retrieve users: %s", err)
	}
	return users, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id int, updateReq *models.User) (*models.User, error) {
	existingUser, err := s.userORM.GetByID(id)
	if err != nil {
		logs.Warn("User not found: %v", err)
		return nil, fmt.Errorf("user not found")
	}

	existingUser.Username = updateReq.Username
	existingUser.Email = updateReq.Email
	existingUser.UpdatedAt = time.Now()

	if err := s.userORM.Update(existingUser); err != nil {
		logs.Error("Failed to update user: %v", err)
		return nil, fmt.Errorf("%s user: %w", constants.ErrFailedToUpdate, err)
	}

	return existingUser, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	if err := s.userORM.Delete(id); err != nil {
		logs.Error("Failed to delete user: %v", err)
		return fmt.Errorf("failed to delete user: %s", err)
	}
	return nil
}

func (s *UserService) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	user, err := s.userORM.GetByID(id)
	if err != nil {
		logs.Warn("User not found: %v", err)
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}
