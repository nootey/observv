package main

import (
	"context"
	"os/signal"
	"syscall"

	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"

	"github.com/nootey/observv/config"
	"github.com/nootey/observv/internal/simulator"
	"github.com/nootey/observv/internal/worker"
	"github.com/nootey/observv/pkg/logging"
)

func main() {

	if err := config.Load(); err != nil {
		panic("failed to load config: " + err.Error())
	}
	cfg := config.Get()

	logger := logging.InitLogger(cfg.App.Environment == "production", cfg.App.LogLevel)
	defer func() {
		_ = logger.Sync()
	}()
	logger = logger.Named("simulator")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	client := simulator.NewClient(cfg.Simulator.APIURL)

	terminals, err := client.WaitForTerminals(ctx, logger)
	if err != nil {
		logger.Info("stopped before the api was ready")
		return
	}
	if len(terminals) < 2 {
		logger.Fatal("transfers need at least two terminals, run the seeder first", zap.Int("terminals", len(terminals)))
	}

	supervisor := worker.NewSupervisor(logger)
	for _, t := range simulator.NewFleet(terminals, client, cfg.Simulator, logger) {
		supervisor.Register(t)
	}

	logger.Info("simulating fleet",
		zap.Int("terminals", len(terminals)),
		zap.Int("interval_seconds", cfg.Simulator.IntervalSeconds),
	)

	if err := supervisor.Run(ctx); err != nil {
		logger.Fatal("supervisor exited with error", zap.Error(err))
	}
	logger.Info("shutdown complete")
}
