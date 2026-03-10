package compaction

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/compaction"
)

type Handler struct {
	web.Controller
	db         *database.Database
	compaction *compaction.Service
}

var compactionService *compaction.Service

func NewHandler(s *compaction.Service) *Handler {
	compactionService = s
	return &Handler{
		db:         s.GetDB(),
		compaction: s,
	}
}

func (h *Handler) Prepare() {
	h.compaction = compactionService
}
