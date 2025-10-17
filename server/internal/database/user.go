package database

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/models"
)

func (db *Database) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := db.ormer.QueryTable(constants.TableNameMap[constants.UserTable]).Filter("username", username).One(&user)
	return &user, err
}

func (db *Database) CompareUserPassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

func (db *Database) CreateUser(user *models.User) error {
	exists := db.ormer.QueryTable(constants.TableNameMap[constants.UserTable]).Filter("username", user.Username).Exist()
	if exists {
		return fmt.Errorf("username already exists")
	}

	_, err := db.ormer.Insert(user)
	return err
}

func (db *Database) ListUsers() ([]*models.User, error) {
	var users []*models.User
	_, err := db.ormer.QueryTable(constants.TableNameMap[constants.UserTable]).All(&users)
	return users, err
}

func (db *Database) GetUserByID(id int) (*models.User, error) {
	user := &models.User{ID: id}
	err := db.ormer.Read(user)
	return user, err
}

func (db *Database) UpdateUser(user *models.User) error {
	_, err := db.ormer.Update(user)
	return err
}

func (db *Database) DeleteUser(id int) error {
	user := &models.User{ID: id}
	_, err := db.ormer.Delete(user)
	return err
}
