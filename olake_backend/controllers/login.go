package controllers

import (
	"encoding/json"
	"net/http"

	// "olake_backend/utils"

	// "olake_backend/utils"

	"olake_backend/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

// LoginController handles authentication requests
type LoginController struct {
	web.Controller
}

// LoginRequest represents the expected JSON structure for login requests
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the JSON structure sent back to the client
type LoginResponse struct {
	Message string `json:"message"` // Human-readable message
	Success bool   `json:"success"` // Indicates if login was successful
}

// Login handles POST requests to /login
// Validates credentials against the database and returns appropriate response
func (c *LoginController) Login() {
	// Initialize the database connection
	// utils.ConnectDB()

	var req LoginRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = LoginResponse{
			Message: "Invalid JSON: " + err.Error(),
			Success: false,
		}
		c.ServeJSON()
		return
	}

	// Use ORM to query the database
	o := orm.NewOrm()
	var user models.User
	err = o.QueryTable("user").Filter("username", req.Username).One(&user)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = LoginResponse{
			Message: "Invalid username or password",
			Success: false,
		}
		c.ServeJSON()
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = LoginResponse{
			Message: "Invalid username or password",
			Success: false,
		}
		c.ServeJSON()
		return
	}

	// On successful login, just return success
	c.Data["json"] = LoginResponse{
		Message: "Login successful",
		Success: true,
	}
	c.ServeJSON()
}

// Add a new method to check authentication
func (c *LoginController) CheckAuth() {
	userID := c.GetSession("user_id")
	if userID == nil {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = LoginResponse{
			Message: "Not authenticated",
			Success: false,
		}
		c.ServeJSON()
		return
	}
	c.Data["json"] = LoginResponse{
		Message: "Authenticated",
		Success: true,
	}
	c.ServeJSON()
}
