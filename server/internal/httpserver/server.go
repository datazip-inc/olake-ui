package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/handlers"
	"github.com/datazip-inc/olake-ui/server/routes"
)

const frontendDistPath = "/opt/frontend/dist"

type Server struct {
	engine     *gin.Engine
	httpServer *http.Server
}

func New(cfg *appconfig.Config, h *handlers.Handler) *Server {
	engine := gin.New()
	engine.Use(gin.Recovery())

	configureMode(cfg)
	configureBaseRoutes(engine)

	if cfg.RunMode == "localdev" {
		engine.Use(localDevCORSMiddleware())
	} else {
		configureStaticFrontend(engine)
	}

	if h != nil {
		routes.RegisterRoutes(engine, h)
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	return &Server{
		engine:     engine,
		httpServer: server,
	}
}

func (s *Server) Engine() *gin.Engine {
	return s.engine
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func configureMode(cfg *appconfig.Config) {
	switch cfg.RunMode {
	case "dev", "localdev":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
}

func configureBaseRoutes(engine *gin.Engine) {
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func configureStaticFrontend(engine *gin.Engine) {
	engine.Static("/assets", filepath.Join(frontendDistPath, "assets"))
	engine.StaticFile("/favicon.ico", filepath.Join(frontendDistPath, "favicon.ico"))
	engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/internal/") {
			c.JSON(http.StatusNotFound, gin.H{"message": "not found", "success": false})
			return
		}
		c.File(filepath.Join(frontendDistPath, "index.html"))
	})
}

func localDevCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type, Accept")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
