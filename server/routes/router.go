package routes

import (
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine, h *handlers.Handler) {
	// core routes
	engine.POST("/signup", h.Signup)
	engine.POST("/login", h.Login)
	engine.GET("/auth/check", h.CheckAuth)
	engine.GET("/telemetry-id", h.TelemetryID)
	engine.GET("/swagger/*any", h.ServeSwagger)

	api := engine.Group("/api")
	api.Use(h.AuthMiddleware())

	etlHandler := h.ETL
	etl := api.Group("/v1")

	// users routes
	etl.POST("/users", etlHandler.CreateUser)
	etl.GET("/users", etlHandler.GetAllUsers)
	etl.PUT("/users/:id", etlHandler.UpdateUser)
	etl.DELETE("/users/:id", etlHandler.DeleteUser)

	// sources routes
	etl.GET("/project/:projectid/sources", etlHandler.ListSources)
	etl.POST("/project/:projectid/sources", etlHandler.CreateSource)
	etl.GET("/project/:projectid/sources/:id", etlHandler.GetSource)
	etl.PUT("/project/:projectid/sources/:id", etlHandler.UpdateSource)
	etl.DELETE("/project/:projectid/sources/:id", etlHandler.DeleteSource)
	etl.POST("/project/:projectid/sources/test", etlHandler.TestSourceConnection)
	etl.POST("/project/:projectid/sources/streams", etlHandler.GetSourceCatalog)
	etl.GET("/project/:projectid/sources/versions", etlHandler.GetSourceVersions)
	etl.POST("/project/:projectid/sources/spec", etlHandler.GetSourceSpec)

	// destinations routes
	etl.GET("/project/:projectid/destinations", etlHandler.ListDestinations)
	etl.POST("/project/:projectid/destinations", etlHandler.CreateDestination)
	etl.PUT("/project/:projectid/destinations/:id", etlHandler.UpdateDestination)
	etl.GET("/project/:projectid/destinations/:id", etlHandler.GetDestination)
	etl.DELETE("/project/:projectid/destinations/:id", etlHandler.DeleteDestination)
	etl.POST("/project/:projectid/destinations/test", etlHandler.TestDestinationConnection)
	etl.GET("/project/:projectid/destinations/versions", etlHandler.GetDestinationVersions)
	etl.POST("/project/:projectid/destinations/spec", etlHandler.GetDestinationSpec)

	// jobs routes
	etl.GET("/project/:projectid/jobs", etlHandler.ListJobs)
	etl.POST("/project/:projectid/jobs", etlHandler.CreateJob)
	etl.GET("/project/:projectid/jobs/:id", etlHandler.GetJob)
	etl.PUT("/project/:projectid/jobs/:id", etlHandler.UpdateJob)
	etl.DELETE("/project/:projectid/jobs/:id", etlHandler.DeleteJob)
	etl.POST("/project/:projectid/jobs/:id/sync", etlHandler.SyncJob)
	etl.POST("/project/:projectid/jobs/:id/activate", etlHandler.ActivateJob)
	etl.GET("/project/:projectid/jobs/:id/tasks", etlHandler.GetJobTasks)
	etl.GET("/project/:projectid/jobs/:id/cancel", etlHandler.CancelJobRun)
	etl.POST("/project/:projectid/jobs/:id/tasks/:taskid/logs", etlHandler.GetTaskLogs)
	etl.GET("/project/:projectid/jobs/:id/logs/download", etlHandler.DownloadTaskLogs)
	etl.POST("/project/:projectid/jobs/:id/clear-destination", etlHandler.ClearDestination)
	etl.GET("/project/:projectid/jobs/:id/clear-destination", etlHandler.GetClearDestinationStatus)
	etl.POST("/project/:projectid/jobs/:id/stream-difference", etlHandler.GetStreamDifference)

	// Project settings routes
	etl.PUT("/project/:projectid/settings", etlHandler.UpsertProjectSettings)
	etl.GET("/project/:projectid/settings", etlHandler.GetProjectSettings)

	// validation routes
	etl.POST("/project/:projectid/check-unique", etlHandler.CheckUniqueName)

	// platform routes
	etl.GET("/platform/releases", etlHandler.GetReleaseUpdates)

	// module gate routes
	etl.GET("/platform/opt/status", h.GetOptimizationStatus)

	// internal routes
	engine.POST("/internal/worker/callback/sync-telemetry", etlHandler.UpdateSyncTelemetry)
	engine.POST("/internal/project/:projectid/jobs/:id/clear-destination/recover", etlHandler.RecoverClearDestination)
	engine.PUT("/internal/project/:projectid/jobs/:id/statefile", etlHandler.UpdateStateFile)

	if h.Optimization != nil {
		optHandler := h.Optimization
		opt := api.Group("/opt/v1")

		// catalogs: crud
		opt.POST("/catalog", optHandler.CreateCatalog)
		opt.GET("/catalog/:catalog", optHandler.GetCatalog)
		opt.PUT("/catalog/:catalog", optHandler.UpdateCatalog)
		opt.DELETE("/catalog/:catalog", optHandler.DeleteCatalog)

		// terminal: cron, enable/disable optimization
		opt.PUT("/:catalog/:database/:table/config", optHandler.SetProperties)

		// tables: view
		opt.GET("/:catalog/:database/tables", optHandler.GetTablesWithDetails)
	}
}

type ModuleNoRouteHandler struct {
	// PathPrefix decides which unmatched paths belong to this module fallback.
	PathPrefix string
	// Middleware is optional and runs before forwarding.
	Middleware gin.HandlerFunc
	// Forward handles the unmatched request (proxy/handler/catch-all).
	Forward gin.HandlerFunc
}

// Module-specific unmatched paths (e.g. /api/opt/v1/*) are delegated here so
// httpserver can keep a single engine-level NoRoute policy.
func HandleModulesNoRoute(c *gin.Context, handlers ...ModuleNoRouteHandler) bool {
	for _, module := range handlers {
		if !strings.HasPrefix(c.Request.URL.Path, module.PathPrefix) {
			continue
		}

		if module.Middleware != nil {
			module.Middleware(c)
			// Aborted means middleware already wrote a response (e.g. 401).
			// Treat it as handled and stop fallback processing.
			if c.IsAborted() {
				return true
			}
		}

		module.Forward(c)
		return true
	}

	return false
}
