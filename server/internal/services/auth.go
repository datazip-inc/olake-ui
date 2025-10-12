package services

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/database"
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
	user, err := s.userORM.FindByUsername(username)
	if err != nil {
		if strings.Contains(err.Error(), "no row found") {
			return nil, constants.ErrUserNotFound
		}
		return nil, fmt.Errorf(constants.ErrFormatFailedToFindUser, err)
	}

	if err := s.userORM.ComparePassword(user.Password, password); err != nil {
		return nil, constants.ErrInvalidCredentials
	}

	telemetry.TrackUserLogin(ctx, user)

	return user, nil
}

func (s *AuthService) Signup(ctx context.Context, user *models.User) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return constants.ErrPasswordProcessing
	}
	user.Password = string(hashedPassword)

	if err := s.userORM.Create(user); err != nil {
		// Check for specific database constraint errors
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return constants.ErrUserAlreadyExists
		}
		return fmt.Errorf("%s user: %w", constants.ErrFailedToCreate, err)
	}
	return nil
}

func (s *AuthService) GetUserByID(userID int) (*models.User, error) {
	return s.userORM.GetByID(userID)
}

func (s *AuthService) ValidateUser(userID int) error {
	_, err := s.userORM.GetByID(userID)
	if err != nil {
		return constants.ErrUserNotFound
	}
	return nil
}
