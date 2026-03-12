package etl

import (
	"context"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/models"
)

// User-related methods on AppService

func (s Service) CreateUser(_ context.Context, req *models.User) error {
	if err := s.db.CreateUser(req); err != nil {
		return fmt.Errorf("failed to create user: %s", err)
	}

	return nil
}

func (s Service) GetAllUsers(_ context.Context) ([]*models.User, error) {
	users, err := s.db.ListUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %s", err)
	}
	return users, nil
}

func (s Service) UpdateUser(_ context.Context, id int, req *models.User) (*models.User, error) {
	existingUser, err := s.db.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %s", err)
	}

	existingUser.Username = req.Username
	existingUser.Email = req.Email

	if err := s.db.UpdateUser(existingUser); err != nil {
		return nil, fmt.Errorf("failed to update user: %s", err)
	}

	return existingUser, nil
}

func (s Service) DeleteUser(_ context.Context, id int) error {
	if err := s.db.DeleteUser(id); err != nil {
		return fmt.Errorf("failed to delete user: %s", err)
	}
	return nil
}

// removed: duplicate of auth.GetUserByID
