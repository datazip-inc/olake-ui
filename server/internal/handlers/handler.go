package handlers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/compaction"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/etl"
	etlservices "github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

type Handler struct {
	web.Controller
	ETL        *etl.Handler
	Compaction *compaction.Handler
}

func NewHandler(etlService *etlservices.ETLService, db *database.Database) *Handler {
	return &Handler{
		ETL:        etl.NewHandler(etlService),
		Compaction: compaction.NewHandler(db),
	}
}
