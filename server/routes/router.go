package routes

import (
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/handlers"
	"github.com/datazip-inc/olake-ui/server/internal/handlers/middleware"
)

// writeDefaultCorsHeaders sets common CORS headers
func writeDefaultCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type, Accept")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

// CustomCorsFilter handles CORS for different route patterns
func CustomCorsFilter(ctx *context.Context) {
	r := ctx.Request
	w := ctx.ResponseWriter
	writeDefaultCorsHeaders(w)
	requestOrigin := r.Header.Get("Origin")
	// API and auth routes - reflect origin
	w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
	// Handle preflight OPTIONS requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
}

func Init(h *handlers.Handler) {
	etlHandler := h.ETL

	if runmode, err := web.AppConfig.String(constants.ConfRunMode); err == nil && runmode == "localdev" {
		web.InsertFilter("*", web.BeforeRouter, CustomCorsFilter)
	} else {
		// Serve static frontend files
		web.SetStaticPath("", "/opt/frontend/dist") // Vite assets are in /assets

		// Serve index.html for React frontend
		web.Router("/*", etlHandler, "get:ServeFrontend") // any other frontend route
	}

	// Swagger routes
	web.Router("/swagger/*", etlHandler, "get:ServeSwagger")

	// Apply auth middleware to protected routes
	web.InsertFilter("/api/*", web.BeforeRouter, middleware.AuthMiddleware)
	// Auth routes
	web.Router("/login", etlHandler, "post:Login")
	web.Router("/signup", etlHandler, "post:Signup")
	web.Router("/auth/check", etlHandler, "get:CheckAuth")
	web.Router("/telemetry-id", etlHandler, "get:GetTelemetryID")

	// User routes
	web.Router("/api/v1/users", etlHandler, "post:CreateUser")
	web.Router("/api/v1/users", etlHandler, "get:GetAllUsers")
	web.Router("/api/v1/users/:id", etlHandler, "put:UpdateUser")
	web.Router("/api/v1/users/:id", etlHandler, "delete:DeleteUser")

	// Source routes
	web.Router("/api/v1/project/:projectid/sources", etlHandler, "get:ListSources")
	web.Router("/api/v1/project/:projectid/sources", etlHandler, "post:CreateSource")
	web.Router("/api/v1/project/:projectid/sources/:id", etlHandler, "get:GetSource")
	web.Router("/api/v1/project/:projectid/sources/:id", etlHandler, "put:UpdateSource")
	web.Router("/api/v1/project/:projectid/sources/:id", etlHandler, "delete:DeleteSource")
	web.Router("/api/v1/project/:projectid/sources/test", etlHandler, "post:TestSourceConnection")
	web.Router("/api/v1/project/:projectid/sources/streams", etlHandler, "post:GetSourceCatalog")
	web.Router("/api/v1/project/:projectid/sources/versions", etlHandler, "get:GetSourceVersions")
	web.Router("/api/v1/project/:projectid/sources/spec", etlHandler, "post:GetSourceSpec")

	// Destination routes
	web.Router("/api/v1/project/:projectid/destinations", etlHandler, "get:ListDestinations")
	web.Router("/api/v1/project/:projectid/destinations", etlHandler, "post:CreateDestination")
	web.Router("/api/v1/project/:projectid/destinations/:id", etlHandler, "get:GetDestination")
	web.Router("/api/v1/project/:projectid/destinations/:id", etlHandler, "put:UpdateDestination")
	web.Router("/api/v1/project/:projectid/destinations/:id", etlHandler, "delete:DeleteDestination")
	web.Router("/api/v1/project/:projectid/destinations/test", etlHandler, "post:TestDestinationConnection")
	web.Router("/api/v1/project/:projectid/destinations/versions", etlHandler, "get:GetDestinationVersions")
	web.Router("/api/v1/project/:projectid/destinations/spec", etlHandler, "post:GetDestinationSpec")

	// Job routes
	web.Router("/api/v1/project/:projectid/jobs", etlHandler, "get:ListJobs")
	web.Router("/api/v1/project/:projectid/jobs", etlHandler, "post:CreateJob")
	web.Router("/api/v1/project/:projectid/jobs/:id", etlHandler, "get:GetJob")
	web.Router("/api/v1/project/:projectid/jobs/:id", etlHandler, "put:UpdateJob")
	web.Router("/api/v1/project/:projectid/jobs/:id", etlHandler, "delete:DeleteJob")
	web.Router("/api/v1/project/:projectid/jobs/:id/sync", etlHandler, "post:SyncJob")
	web.Router("/api/v1/project/:projectid/jobs/:id/activate", etlHandler, "post:ActivateJob")
	web.Router("/api/v1/project/:projectid/jobs/:id/tasks", etlHandler, "get:GetJobTasks")
	web.Router("/api/v1/project/:projectid/jobs/:id/cancel", etlHandler, "get:CancelJobRun")
	web.Router("/api/v1/project/:projectid/jobs/:id/tasks/:taskid/logs", etlHandler, "post:GetTaskLogs")
	web.Router("/api/v1/project/:projectid/jobs/:id/logs/download", etlHandler, "get:DownloadTaskLogs")
	web.Router("/api/v1/project/:projectid/jobs/:id/clear-destination", etlHandler, "post:ClearDestination")
	web.Router("/api/v1/project/:projectid/jobs/:id/clear-destination", etlHandler, "get:GetClearDestinationStatus")
	web.Router("/api/v1/project/:projectid/jobs/:id/stream-difference", etlHandler, "post:GetStreamDifference")

	// Project settings routes
	web.Router("/api/v1/project/:projectid/settings", etlHandler, "put:UpsertProjectSettings")
	web.Router("/api/v1/project/:projectid/settings", etlHandler, "get:GetProjectSettings")

	// validation routes
	web.Router("/api/v1/project/:projectid/check-unique", etlHandler, "post:CheckUniqueName")

	// platform routes
	web.Router("/api/v1/platform/releases", etlHandler, "get:GetReleaseUpdates")
	web.Router("/api/v1/platform/opt/status", h, "get:GetoptimizationStatus")

	// internal routes
	web.Router("/internal/worker/callback/sync-telemetry", etlHandler, "post:UpdateSyncTelemetry")
	web.Router("/internal/project/:projectid/jobs/:id/clear-destination/recover", etlHandler, "post:RecoverClearDestination")
	web.Router("/internal/project/:projectid/jobs/:id/statefile", etlHandler, "put:UpdateStateFile")

	if h.Optimization != nil {
		optHandler := h.Optimization

		// catalogs: crud
		web.Router("/api/opt/v1/catalog", optHandler, "post:CreateCatalog")
		web.Router("/api/opt/v1/catalog/:catalog", optHandler, "get:GetCatalog")
		web.Router("/api/opt/v1/catalog/:catalog", optHandler, "put:UpdateCatalog")
		web.Router("/api/opt/v1/catalog/:catalog", optHandler, "delete:DeleteCatalog")

		// terminal: cron, enable/disable optimization
		web.Router("/api/opt/v1/:catalog/:database/:table/config", optHandler, "put:SetProperties")

		// tables: view
		web.Router("/api/opt/v1/:catalog/:database/tables", optHandler, "get:GetTablesWithDetails")

		// piggy backing
		web.Router("/api/opt/v1/*", optHandler, "get:PiggyBacking;post:PiggyBacking;put:PiggyBacking;delete:PiggyBacking")
	}
}
