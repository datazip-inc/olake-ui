package handlers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/services"
)

type Handler struct {
	web.Controller
	svc *services.AppService
}

func NewHandler(s *services.AppService) *Handler {
	return &Handler{svc: s}
}
