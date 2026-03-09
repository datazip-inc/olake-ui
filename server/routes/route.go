package routes

import (
	"github.com/datazip-inc/olake-ui/server/internal/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine, h *handlers.Handler) {
	engine.POST("/login", h.Login)
	engine.POST("/signup", h.Signup)
	engine.GET("/auth/check", h.CheckAuth)
	engine.GET("/telemetry-id", h.TelemetryID)
	engine.GET("/swagger/*any", h.ServeSwagger)

	v1 := engine.Group("/api/v1")
	v1.Use(h.AuthMiddleware())

	{
		// users routes
		v1.POST("/users", h.CreateUser)
		v1.GET("/users", h.GetAllUsers)
		v1.PUT("/users/:id", h.UpdateUser)
		v1.DELETE("/users/:id", h.DeleteUser)

		// sources routes
		v1.GET("/project/:projectid/sources", h.ListSources)
		v1.POST("/project/:projectid/sources", h.CreateSource)
		v1.GET("/project/:projectid/sources/:id", h.GetSource)
		v1.PUT("/project/:projectid/sources/:id", h.UpdateSource)
		v1.DELETE("/project/:projectid/sources/:id", h.DeleteSource)
		v1.POST("/project/:projectid/sources/test", h.TestSourceConnection)
		v1.POST("/project/:projectid/sources/streams", h.GetSourceCatalog)
		v1.GET("/project/:projectid/sources/versions", h.GetSourceVersions)
		v1.POST("/project/:projectid/sources/spec", h.GetSourceSpec)

		// destinations routes
		v1.GET("/project/:projectid/destinations", h.ListDestinations)
		v1.POST("/project/:projectid/destinations", h.CreateDestination)
		v1.GET("/project/:projectid/destinations/:id", h.GetDestination)
		v1.PUT("/project/:projectid/destinations/:id", h.UpdateDestination)
		v1.DELETE("/project/:projectid/destinations/:id", h.DeleteDestination)
		v1.POST("/project/:projectid/destinations/test", h.TestDestinationConnection)
		v1.GET("/project/:projectid/destinations/versions", h.GetDestinationVersions)
		v1.POST("/project/:projectid/destinations/spec", h.GetDestinationSpec)

		// jobs routes
		v1.GET("/project/:projectid/jobs", h.ListJobs)
		v1.POST("/project/:projectid/jobs", h.CreateJob)
		v1.GET("/project/:projectid/jobs/:id", h.GetJob)
		v1.PUT("/project/:projectid/jobs/:id", h.UpdateJob)
		v1.DELETE("/project/:projectid/jobs/:id", h.DeleteJob)
		v1.POST("/project/:projectid/jobs/:id/sync", h.SyncJob)
		v1.POST("/project/:projectid/jobs/:id/activate", h.ActivateJob)
		v1.GET("/project/:projectid/jobs/:id/tasks", h.GetJobTasks)
		v1.GET("/project/:projectid/jobs/:id/cancel", h.CancelJobRun)
		v1.POST("/project/:projectid/jobs/:id/tasks/:taskid/logs", h.GetTaskLogs)
		v1.GET("/project/:projectid/jobs/:id/logs/download", h.DownloadTaskLogs)
		v1.POST("/project/:projectid/jobs/:id/clear-destination", h.ClearDestination)
		v1.GET("/project/:projectid/jobs/:id/clear-destination", h.GetClearDestinationStatus)
		v1.POST("/project/:projectid/jobs/:id/stream-difference", h.GetStreamDifference)

		v1.PUT("/project/:projectid/settings", h.UpsertProjectSettings)
		v1.GET("/project/:projectid/settings", h.GetProjectSettings)
		v1.POST("/project/:projectid/check-unique", h.CheckUniqueName)
		v1.GET("/platform/releases", h.GetReleaseUpdates)
	}

	// internal routes
	engine.POST("/internal/worker/callback/sync-telemetry", h.UpdateSyncTelemetry)
	engine.POST("/internal/project/:projectid/jobs/:id/clear-destination/recover", h.RecoverClearDestination)
	engine.PUT("/internal/project/:projectid/jobs/:id/statefile", h.UpdateStateFile)
}
