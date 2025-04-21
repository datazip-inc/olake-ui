package controllers

import (
	"encoding/json"
	"net/http"
	"olake_backend/models"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	web.Controller
	o orm.Ormer
}

// Prepare initializes the ORM instance
func (c *UserController) Prepare() {
	c.o = orm.NewOrm() // Initialize the ORM instance
}

// CreateUser handles the creation of a new user
func (c *UserController) CreateUser() {
	var user models.User
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &user); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	// Save user to the database (pseudo code)
	// db.Create(&user)
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = user
	c.ServeJSON()
}

// GetAllUsers retrieves all users
func (c *UserController) GetAllUsers() {
	var users []models.User
	// Retrieve users from the database (pseudo code)
	// db.Find(&users)
	c.Data["json"] = users
	c.ServeJSON()
}

// UpdateUser updates an existing user
func (c *UserController) UpdateUser() {
	id := c.Ctx.Input.Param(":id")
	var user models.User
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &user); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	// Use id to find the user in the database
	_, err := c.o.QueryTable("user").Filter("user_id", id).Update(orm.Params{
		"name":      user.Username,
		"email":     user.Email,
		"updated_at": time.Now(),
	})
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to update user"}
		c.ServeJSON()
		return
	}
	c.Data["json"] = user
	c.ServeJSON()
}

// DeleteUser deletes a user
func (c *UserController) DeleteUser() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseUint(idStr, 10, 32) // Convert string to uint
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": "Invalid ID"}
		c.ServeJSON()
		return
	}
	_, err = c.o.Delete(&models.User{UserID: int(id)}) // Use the converted id
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to delete user"}
		c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusNoContent)
}

// Signup handles user registration
func (c *UserController) Signup() {
	var user models.User
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &user); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to hash password"}
		c.ServeJSON()
		return
	}
	user.Password = string(hashedPassword)

	// Set CreatedAt and UpdatedAt to the current time
	currentTime := time.Now()
	user.CreatedAt = currentTime
	user.UpdatedAt = currentTime

	// Save user to the database
	_, err = c.o.Insert(&user)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]string{"error": "Failed to create user"}
		c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = map[string]string{"message": "User created successfully. Please log in."}
	c.ServeJSON()
} 