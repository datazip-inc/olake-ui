package etl

import (
	"context"
	"errors"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/utils/telemetry"
	"golang.org/x/crypto/bcrypt"
)

// Auth-related methods on AppService

func (s Service) Login(ctx context.Context, username, password string) (*models.User, error) {
	user, err := s.db.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, constants.ErrUserNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get user: %s", err)
	}

	if err := s.db.CompareUserPassword(user.Password, password); err != nil {
		return nil, fmt.Errorf("%w: %v", constants.ErrInvalidCredentials, err)
	}

	telemetry.TrackUserLogin(ctx, user)

	return user, nil
}

func (s *Service) Signup(_ context.Context, req *dto.CreateUserRequest) error {
	user := &models.User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%w: %v", constants.ErrPasswordProcessing, err)
	}
	user.Password = string(hashedPassword)

	if err := s.db.CreateUser(user); err != nil {
		if errors.Is(err, constants.ErrUserAlreadyExists) {
			return err
		}
		return fmt.Errorf("failed to create user: %s", err)
	}

	return nil
}

func (s Service) GetUserByID(userID int) (*models.User, error) {
	user, err := s.db.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, constants.ErrUserNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to find user: %s", err)
	}
	return user, nil
}

func (s Service) ValidateUser(userID int) error {
	_, err := s.db.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, constants.ErrUserNotFound) {
			return err
		}
		return fmt.Errorf("failed to validate user: %s", err)
	}
	return nil
}
