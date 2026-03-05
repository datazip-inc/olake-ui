package compaction

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/database"
)

type Handler struct {
	web.Controller
	db *database.Database
}

func NewHandler(db *database.Database) *Handler {
	return &Handler{
		db: db,
	}
}

func (h *Handler) Prepare() {
}
