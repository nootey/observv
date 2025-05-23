package runtime

import (
	"context"
	"fmt"
	"github.com/nootey/observv/internal/bootstrap"
	"github.com/nootey/observv/internal/http"
	"github.com/nootey/observv/internal/simulate"
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

func (rt *HTTPServerRuntime) Run(context context.Context, simulateTerminalTraffic bool) error {
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

	container := bootstrap.NewContainer(rt.Config, dbClient)
	httpLogger := rt.Logger.Named("http").With(zap.String("component", "HTTP"))
	httpServer := http.NewServer(container, httpLogger)
	go httpServer.Start()

	// Once DB and HTTP are confirmed healthy, run the simulation
	if simulateTerminalTraffic {
		rt.Logger.Info("Simulation enabled: starting fake terminal traffic...")
		simulationLogger := rt.Logger.Named("simulation")
		go simulate.StartFakeTraffic(container, simulationLogger)
		go simulate.StartStatusToggleLoop(container, simulationLogger)
		go simulate.StartUptimeTracker(container, simulationLogger)
	}

	<-ctx.Done()

	rt.Logger.Info("Interrupt signal received, shutting down HTTP server...")
	if err := httpServer.Shutdown(); err != nil {
		rt.Logger.Error("HTTP server shutdown failed", zap.Error(err))
		return fmt.Errorf("http server shutdown failed: %w", err)
	}
	rt.Logger.Info("HTTP server exiting")
	return nil
}
