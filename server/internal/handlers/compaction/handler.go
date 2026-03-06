package compaction

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	services "github.com/datazip-inc/olake-ui/server/internal/services/compaction"
)

type Handler struct {
	web.Controller
	db         *database.Database
	compaction *services.CompactionService
}

var compactionService *services.CompactionService

func NewHandler(s *services.CompactionService) *Handler {
	compactionService = s
	return &Handler{
		db:         s.GetDB(),
		compaction: s,
	}
}

func (h *Handler) Prepare() {
	h.compaction = compactionService
}
