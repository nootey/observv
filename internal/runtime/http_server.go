package runtime

import (
	"context"
	"fmt"
	"github.com/nootey/observv/internal/bootstrap"
	"github.com/nootey/observv/internal/http"
	"github.com/nootey/observv/pkg/config"
	"github.com/nootey/observv/pkg/database"
	"go.uber.org/zap"
	"os/signal"
	"syscall"
)

type HTTPServerRuntime struct {
	Logger *zap.Logger
	Config *config.Config
}

func NewHTTPServerRuntime(cfg *config.Config, logger *zap.Logger) *HTTPServerRuntime {
	return &HTTPServerRuntime{
		Config: cfg,
		Logger: logger,
	}
}

func (rt *HTTPServerRuntime) Run(context context.Context) error {
	ctx, stop := signal.NotifyContext(context, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := database.EnsureDatabaseExists(rt.Config); err != nil {
		rt.Logger.Error("Database creation failed", zap.Error(err))
		return fmt.Errorf("failed to ensure database exists: %w", err)
	}

	dbClient, err := database.ConnectToDatabase(rt.Config, true)
	if err != nil {
		rt.Logger.Error("Database connection failed", zap.Error(err))
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer database.DisconnectDatabase()
	rt.Logger.Info("Successfully connected to the database")

	err = database.MigrateAndSeed(rt.Config, dbClient, rt.Logger)
	if err != nil {
		rt.Logger.Error("Initial database migration failed", zap.Error(err))
		return fmt.Errorf("initial database migration failed: %w", err)
	}

	container := bootstrap.NewContainer(rt.Config, dbClient)
	httpLogger := rt.Logger.Named("http").With(zap.String("component", "HTTP"))
	httpServer := http.NewServer(container, httpLogger)
	go httpServer.Start()

	<-ctx.Done()

	rt.Logger.Info("Interrupt signal received, shutting down HTTP server...")
	if err := httpServer.Shutdown(); err != nil {
		rt.Logger.Error("HTTP server shutdown failed", zap.Error(err))
		return fmt.Errorf("http server shutdown failed: %w", err)
	}
	rt.Logger.Info("HTTP server exiting")
	return nil
}
