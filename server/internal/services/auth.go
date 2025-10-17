package services

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/telemetry"
)

// Auth-related methods on AppService

func (s *AppService) Login(ctx context.Context, username, password string) (*models.User, error) {
	user, err := s.db.GetUserByUsername(username)
	if err != nil {
		if strings.Contains(err.Error(), "no row found") {
			return nil, fmt.Errorf("user not found: %s", err)
		}
		return nil, fmt.Errorf("failed to get user: %s", err)
	}

	if err := s.db.CompareUserPassword(user.Password, password); err != nil {
		return nil, fmt.Errorf("invalid credentials: %s", err)
	}

	telemetry.TrackUserLogin(ctx, user)

	return user, nil
}

func (s *AppService) Signup(_ context.Context, user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %s", err)
	}
	user.Password = string(hashedPassword)

	if err := s.db.CreateUser(user); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return fmt.Errorf("user already exists: %s", err)
		}
		return fmt.Errorf("failed to create user: %s", err)
	}

	return nil
}

func (s *AppService) GetUserByID(userID int) (*models.User, error) {
	user, err := s.db.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %s", err)
	}
	return user, nil
}

func (s *AppService) ValidateUser(userID int) error {
	_, err := s.db.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to validate user: %s", err)
	}
	return nil
}
