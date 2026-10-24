package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"

	"github.com/nootey/observv/config"
	"github.com/nootey/observv/internal/simulator"
	"github.com/nootey/observv/internal/worker"
	"github.com/nootey/observv/pkg/logging"
	"github.com/nootey/observv/pkg/telemetry"
	"github.com/nootey/observv/pkg/version"
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

	tel, err := telemetry.New(context.Background(), telemetry.Config{
		ServiceName: "observv-simulator",
		Version:     version.Version,
		Environment: cfg.App.Environment,
		Endpoint:    cfg.OTel.ExporterOTLPEndpoint,
		Disabled:    cfg.OTel.SDKDisabled,
	}, logger.Named("telemetry"))
	if err != nil {
		logger.Fatal("failed to start telemetry", zap.Error(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tel.Shutdown(ctx); err != nil {
			logger.Error("telemetry shutdown failed", zap.Error(err))
		}
	}()

	logger = logging.WithOTel(logger, tel.LoggerProvider()).Named("simulator")

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
