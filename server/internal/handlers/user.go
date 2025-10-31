package handlers

import (
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/constants"
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
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	if err := c.userService.CreateUser(c.Ctx.Request.Context(), &req); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to create user", err)
		return
	}
	utils.SuccessResponse(&c.Controller, req)
}

// @router /users [get]
func (c *UserHandler) GetAllUsers() {
	users, err := c.userService.GetAllUsers(c.Ctx.Request.Context())
	if err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to get users", err)
		return
	}
	utils.SuccessResponse(&c.Controller, users)
}

// @router /users/:id [put]
func (c *UserHandler) UpdateUser() {
	id := GetIDFromPath(&c.Controller)

	var req models.User
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	updatedUser, err := c.userService.UpdateUser(c.Ctx.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "user not found" {
			respondWithError(&c.Controller, http.StatusNotFound, "User not found", err)
			return
		}
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to update user", err)
		return
	}
	utils.SuccessResponse(&c.Controller, updatedUser)
}

// @router /users/:id [delete]
func (c *UserHandler) DeleteUser() {
	id := GetIDFromPath(&c.Controller)
	if err := c.userService.DeleteUser(c.Ctx.Request.Context(), id); err != nil {
		respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to delete user", err)
		return
	}

	utils.SuccessResponse(&c.Controller, nil)
}
