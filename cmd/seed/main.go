package main

import (
	"context"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"

	"github.com/nootey/observv/config"
	"github.com/nootey/observv/internal/store"
	"github.com/nootey/observv/internal/store/seeders"
	"github.com/nootey/observv/pkg/logging"
)

func main() {

	if err := config.Load(); err != nil {
		panic("failed to load config: " + err.Error())
	}
	cfg := config.Get()

	logger := logging.InitLogger(cfg.App.Environment == "production", cfg.App.LogLevel)
	defer func(log *zap.Logger) {
		_ = log.Sync() // Ignore sync errors
	}(logger)
	logger = logger.Named("seeder")

	db, err := store.NewDB(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database connection", zap.Error(err))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	logger.Info("starting database seeding")

	if err := seeders.SeedDatabase(ctx, db, logger, cfg); err != nil {
		logger.Fatal("seeding failed", zap.Error(err))
	}

	logger.Info("seeding completed successfully")
}
