package http

import (
	"context"
	"errors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/nootey/observv/internal/bootstrap"
	"github.com/nootey/observv/internal/middleware"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Server struct {
	Router *gin.Engine
	server *http.Server
	logger *zap.Logger
}

func NewServer(container *bootstrap.Container, logger *zap.Logger) *Server {

	router := NewRouter(container)

	addr := container.Config.Host + ":" + container.Config.HttpPort

	return &Server{
		Router: router,
		logger: logger,
		server: &http.Server{
			Addr: addr,
		},
	}
}

func (s *Server) Start() {
	s.logger.Info("Starting the HTTP server")

	s.Router.Use(ginzap.Ginzap(s.logger, time.RFC3339, true), ginzap.RecoveryWithZap(s.logger, true))

	s.server.Handler = s.Router.Handler()

	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatal("Failed to listen and serve for: HTTP", zap.Error(err))
		}
	}()
}

func (s *Server) Shutdown() error {
	s.logger.Info("Shutting down the HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func NewRouter(container *bootstrap.Container) *gin.Engine {

	var r *gin.Engine
	r = gin.Default()

	r.Use(gin.Recovery())

	// Register middleware
	r.Use(gin.Recovery())
	r.Use(middleware.PrometheusMiddleware())

	// Prometheus endpoint
	prometheus := ginprometheus.NewPrometheus("http")
	prometheus.Use(r)

	// Initialize routes
	routeInitializer := NewRouteInitializerHTTP(r, container)
	routeInitializer.InitEndpoints()

	return r
}
