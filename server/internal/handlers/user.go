package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/services"
	"github.com/datazip/olake-ui/server/utils"
)

type UserHandler struct {
	web.Controller
	userService *services.UserService
}

func (c *UserHandler) Prepare() {
	c.userService = services.NewUserService()
}

// @router /users [post]
func (c *UserHandler) CreateUser() {
	var req models.User
	if err := bindJSON(&c.Controller, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := c.userService.CreateUser(context.Background(), &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /users [get]
func (c *UserHandler) GetAllUsers() {
	users, err := c.userService.GetAllUsers(context.Background())
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
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
	if err := bindJSON(&c.Controller, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	updatedUser, err := c.userService.UpdateUser(context.Background(), id, &req)
	if err != nil {
		if err.Error() == "user not found" {
			utils.ErrorResponse(&c.Controller, http.StatusNotFound, "User not found")
			return
		}
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(&c.Controller, updatedUser)
}

// @router /users/:id [delete]
func (c *UserHandler) DeleteUser() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := c.userService.DeleteUser(context.Background(), id); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, err.Error())
		return
	}

	c.Ctx.ResponseWriter.WriteHeader(http.StatusNoContent)
}
