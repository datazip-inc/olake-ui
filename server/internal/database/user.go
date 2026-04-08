package database

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
)

func (db *Database) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := db.conn.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %v", constants.ErrUserNotFound, err)
		}
		return nil, err
	}
	return &user, nil
}

func (db *Database) CompareUserPassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

func (db *Database) CreateUser(user *models.User) error {
	var count int64
	if err := db.conn.Model(&models.User{}).Where("username = ?", user.Username).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w", constants.ErrUserAlreadyExists)
	}
	if err := db.conn.Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (db *Database) ListUsers() ([]*models.User, error) {
	var users []*models.User
	err := db.conn.Find(&users).Error
	return users, err
}

func (db *Database) GetUserByID(id int) (*models.User, error) {
	user := &models.User{}
	err := db.conn.First(user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: user not found id[%d]", constants.ErrUserNotFound, id)
		}
		return nil, err
	}
	return user, nil
}

func (db *Database) UpdateUser(user *models.User) error {
	return db.conn.Updates(user).Error
}

func (db *Database) DeleteUser(id int) error {
	result := db.conn.Delete(&models.User{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return constants.ErrUserNotFound
	}
	return nil
}
