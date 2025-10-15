package database

import (
	"fmt"

	"github.com/beego/beego/v2/client/orm"
	"golang.org/x/crypto/bcrypt"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/models"
)

// UserORM handles database operations
type UserORM struct {
	ormer     orm.Ormer
	TableName string
}

func NewUserORM() *UserORM {
	return &UserORM{
		ormer:     orm.NewOrm(),
		TableName: constants.TableNameMap[constants.UserTable],
	}
}

func (r *UserORM) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.ormer.QueryTable(r.TableName).Filter("username", username).One(&user)
	return &user, err
}

func (r *UserORM) ComparePassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

func (r *UserORM) Create(user *models.User) error {
	exists := r.ormer.QueryTable(r.TableName).Filter("username", user.Username).Exist()
	if exists {
		return fmt.Errorf("username already exists")
	}

	_, err := r.ormer.Insert(user)
	return err
}

func (r *UserORM) GetAll() ([]*models.User, error) {
	var users []*models.User
	_, err := r.ormer.QueryTable(r.TableName).All(&users)
	return users, err
}

func (r *UserORM) GetByID(id int) (*models.User, error) {
	user := &models.User{ID: id}
	err := r.ormer.Read(user)
	return user, err
}

func (r *UserORM) Update(user *models.User) error {
	_, err := r.ormer.Update(user)
	return err
}

func (r *UserORM) Delete(id int) error {
	user := &models.User{ID: id}
	_, err := r.ormer.Delete(user)
	return err
}
