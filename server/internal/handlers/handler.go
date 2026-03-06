package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	services "github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

type GinHandler struct {
	etl      *services.ETLService
	sessions *sessionStore
}

func NewGinHandler(s *services.ETLService, cfg appconfig.Config) (*GinHandler, error) {
	sessionStore, err := newSessionStore(cfg)
	if err != nil {
		return nil, err
	}
	return &GinHandler{
		etl:      s,
		sessions: sessionStore,
	}, nil
}

func (h *GinHandler) RegisterRoutes(engine *gin.Engine) {
	engine.POST("/login", h.login)
	engine.POST("/signup", h.signup)
	engine.GET("/auth/check", h.checkAuth)
	engine.GET("/telemetry-id", h.telemetryID)
	engine.GET("/swagger/*any", h.serveSwagger)

	v1 := engine.Group("/api/v1")
	v1.Use(h.authMiddleware())

	{
		// users routes
		v1.POST("/users", h.createUser)
		v1.GET("/users", h.getAllUsers)
		v1.PUT("/users/:id", h.updateUser)
		v1.DELETE("/users/:id", h.deleteUser)

		// sources routes
		v1.GET("/project/:projectid/sources", h.listSources)
		v1.POST("/project/:projectid/sources", h.createSource)
		v1.GET("/project/:projectid/sources/:id", h.getSource)
		v1.PUT("/project/:projectid/sources/:id", h.updateSource)
		v1.DELETE("/project/:projectid/sources/:id", h.deleteSource)
		v1.POST("/project/:projectid/sources/test", h.testSourceConnection)
		v1.POST("/project/:projectid/sources/streams", h.getSourceCatalog)
		v1.GET("/project/:projectid/sources/versions", h.getSourceVersions)
		v1.POST("/project/:projectid/sources/spec", h.getSourceSpec)

		// destinations routes
		v1.GET("/project/:projectid/destinations", h.listDestinations)
		v1.POST("/project/:projectid/destinations", h.createDestination)
		v1.GET("/project/:projectid/destinations/:id", h.getDestination)
		v1.PUT("/project/:projectid/destinations/:id", h.updateDestination)
		v1.DELETE("/project/:projectid/destinations/:id", h.deleteDestination)
		v1.POST("/project/:projectid/destinations/test", h.testDestinationConnection)
		v1.GET("/project/:projectid/destinations/versions", h.getDestinationVersions)
		v1.POST("/project/:projectid/destinations/spec", h.getDestinationSpec)

		// jobs routes
		v1.GET("/project/:projectid/jobs", h.listJobs)
		v1.POST("/project/:projectid/jobs", h.createJob)
		v1.GET("/project/:projectid/jobs/:id", h.getJob)
		v1.PUT("/project/:projectid/jobs/:id", h.updateJob)
		v1.DELETE("/project/:projectid/jobs/:id", h.deleteJob)
		v1.POST("/project/:projectid/jobs/:id/sync", h.syncJob)
		v1.POST("/project/:projectid/jobs/:id/activate", h.activateJob)
		v1.GET("/project/:projectid/jobs/:id/tasks", h.getJobTasks)
		v1.GET("/project/:projectid/jobs/:id/cancel", h.cancelJobRun)
		v1.POST("/project/:projectid/jobs/:id/tasks/:taskid/logs", h.getTaskLogs)
		v1.GET("/project/:projectid/jobs/:id/logs/download", h.downloadTaskLogs)
		v1.POST("/project/:projectid/jobs/:id/clear-destination", h.clearDestination)
		v1.GET("/project/:projectid/jobs/:id/clear-destination", h.getClearDestinationStatus)
		v1.POST("/project/:projectid/jobs/:id/stream-difference", h.getStreamDifference)

		v1.PUT("/project/:projectid/settings", h.upsertProjectSettings)
		v1.GET("/project/:projectid/settings", h.getProjectSettings)
		v1.POST("/project/:projectid/check-unique", h.checkUniqueName)
		v1.GET("/platform/releases", h.getReleaseUpdates)
	}

	// internal routes
	engine.POST("/internal/worker/callback/sync-telemetry", h.updateSyncTelemetry)
	engine.POST("/internal/project/:projectid/jobs/:id/clear-destination/recover", h.recoverClearDestination)
	engine.PUT("/internal/project/:projectid/jobs/:id/statefile", h.updateStateFile)
}
