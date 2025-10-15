package services

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/database"
	"github.com/datazip/olake-ui/server/internal/dto"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/telemetry"
)

type AuthService struct {
	userORM *database.UserORM
}

func NewAuthService() *AuthService {
	return &AuthService{
		userORM: database.NewUserORM(),
	}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*models.User, error) {
	if err := dto.Validate(&models.User{Username: username, Password: password}); err != nil {
		return nil, fmt.Errorf("invalid credentials - username=%s error=%v: %w", username, err, constants.ErrInvalidCredentials)
	}

	user, err := s.userORM.FindByUsername(username)
	if err != nil {
		if strings.Contains(err.Error(), "no row found") {
			return nil, fmt.Errorf("user not found - username=%s error=%v: %w", username, err, constants.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to find user - username=%s error=%v", username, err)
	}

	if err := s.userORM.ComparePassword(user.Password, password); err != nil {
		return nil, fmt.Errorf("invalid credentials - username=%s error=%v: %w", username, err, constants.ErrInvalidCredentials)
	}

	telemetry.TrackUserLogin(ctx, user)

	return user, nil
}

func (s *AuthService) Signup(_ context.Context, user *models.User) error {
	if err := dto.Validate(user); err != nil {
		return fmt.Errorf("failed to validate signup request - username=%s email=%s error=%v: %w",
			user.Username, user.Email, err, constants.ErrInvalidCredentials)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password - username=%s error=%v: %w", user.Username, err, constants.ErrPasswordProcessing)
	}
	user.Password = string(hashedPassword)

	if err := s.userORM.Create(user); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return fmt.Errorf("user already exists - username=%s email=%s error=%v: %w",
				user.Username, user.Email, err, constants.ErrUserAlreadyExists)
		}
		return fmt.Errorf("failed to create user - username=%s email=%s error=%v", user.Username, user.Email, err)
	}

	return nil
}

func (s *AuthService) GetUserByID(userID int) (*models.User, error) {
	user, err := s.userORM.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user - user_id=%d error=%v", userID, err)
	}
	return user, nil
}

func (s *AuthService) ValidateUser(userID int) error {
	_, err := s.userORM.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to validate user - user_id=%d error=%v", userID, err)
	}
	return nil
}
