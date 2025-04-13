package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/beego/beego/v2/server/web"

	"github.com/datazip/olake-server/internal/database"
	"github.com/datazip/olake-server/internal/models"
	"github.com/datazip/olake-server/utils"
)

type UserHandler struct {
	web.Controller
	userORM *database.UserORM
}

func (c *UserHandler) Prepare() {
	c.userORM = database.NewUserORM()
}

// @router /users [post]
func (c *UserHandler) CreateUser() {
	var req models.User
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := c.userORM.Create(&req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to create user: %s", err))
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /users [get]
func (c *UserHandler) GetAllUsers() {
	users, err := c.userORM.GetAll()
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}

	utils.SuccessResponse(&c.Controller, users)
}

// @router /users/:id [put]
func (c *UserHandler) UpdateUser() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req models.User
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Get existing user
	existingUser, err := c.userORM.GetByID(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusNotFound, "User not found")
		return
	}

	// Update fields
	existingUser.Username = req.Username
	existingUser.Email = req.Email
	existingUser.UpdatedAt = time.Now()

	if err := c.userORM.Update(existingUser); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update user")
		return
	}

	utils.SuccessResponse(&c.Controller, existingUser)
}

// @router /users/:id [delete]
func (c *UserHandler) DeleteUser() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := c.userORM.Delete(id); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	c.Ctx.ResponseWriter.WriteHeader(http.StatusNoContent)
}
