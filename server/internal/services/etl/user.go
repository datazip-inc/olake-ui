package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// User-related methods on AppService

func (s *ETLService) CreateUser(_ context.Context, req *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%w: %v", constants.ErrPasswordProcessing, err)
	}
	req.Password = string(hashedPassword)

	if err := s.db.CreateUser(req); err != nil {
		if errors.Is(err, constants.ErrUserAlreadyExists) {
			return err
		}
		return fmt.Errorf("failed to create user: %s", err)
	}

	return nil
}

func (s *ETLService) GetAllUsers(_ context.Context) ([]*models.User, error) {
	users, err := s.db.ListUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %s", err)
	}
	return users, nil
}

func (s *ETLService) UpdateUser(_ context.Context, id int, req *models.User) (*models.User, error) {
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

func (s *ETLService) DeleteUser(_ context.Context, id int) error {
	if err := s.db.DeleteUser(id); err != nil {
		return fmt.Errorf("failed to delete user: %s", err)
	}
	return nil
}

// removed: duplicate of auth.GetUserByID
