package services

import (
	"context"
	"fmt"

	"github.com/datazip/olake-ui/server/internal/database"
	"github.com/datazip/olake-ui/server/internal/dto"
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
	if err := dto.Validate(&req); err != nil {
		return fmt.Errorf("failed to validate user request - username=%s email=%s error=%v",
			req.Username, req.Email, err)
	}

	if err := s.userORM.Create(req); err != nil {
		return fmt.Errorf("failed to create user - username=%s email=%s error=%v", req.Username, req.Email, err)
	}

	return nil
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	users, err := s.userORM.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve users - error=%v", err)
	}
	return users, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id int, req *models.User) (*models.User, error) {
	if err := dto.Validate(&req); err != nil {
		return nil, fmt.Errorf("failed to validate update user request - user_id=%d username=%s error=%v",
			id, req.Username, err)
	}

	existingUser, err := s.userORM.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found - user_id=%d error=%v", id, err)
	}

	existingUser.Username = req.Username
	existingUser.Email = req.Email

	if err := s.userORM.Update(existingUser); err != nil {
		return nil, fmt.Errorf("failed to update user - user_id=%d username=%s error=%v", id, req.Username, err)
	}

	return existingUser, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	if err := s.userORM.Delete(id); err != nil {
		return fmt.Errorf("failed to delete user - user_id=%d error=%v", id, err)
	}
	return nil
}

func (s *UserService) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	user, err := s.userORM.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found - user_id=%d error=%v", id, err)
	}
	return user, nil
}
