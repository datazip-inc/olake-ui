package services

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/models"
)

type AuthService struct {
	userORM *database.UserORM
}

func NewAuthService() *AuthService {
	return &AuthService{
		userORM: database.NewUserORM(),
	}
}

func (s *AuthService) Login(username, password string) (*models.User, error) {
	user, err := s.userORM.FindByUsername(username)
	if err != nil {
		if strings.Contains(err.Error(), "no row found") {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf(ErrFormatFailedToFindUser, err)
	}

	if err := s.userORM.ComparePassword(user.Password, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (s *AuthService) Signup(user *models.User) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return ErrPasswordProcessing
	}
	user.Password = string(hashedPassword)

	if err := s.userORM.Create(user); err != nil {
		// Check for specific database constraint errors
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("%s user: %w", ErrFailedToCreate, err)
	}

	return nil
}

func (s *AuthService) GetUserByID(userID int) (*models.User, error) {
	return s.userORM.GetByID(userID)
}

func (s *AuthService) ValidateUser(userID int) error {
	_, err := s.userORM.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}
	return nil
}
