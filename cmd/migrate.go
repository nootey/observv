package main

import (
	"fmt"
	"github.com/nootey/observv/internal/models"
	"github.com/nootey/observv/pkg/config"
	"github.com/nootey/observv/pkg/database"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate [fresh]",
	Short: "Run database migrations. Use 'fresh' to drop and recreate the database before migrating.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fresh := len(args) == 1 && args[0] == "fresh"

		ctx := cmd.Context()
		logger := ctx.Value("logger").(*zap.Logger)

		return runMigrations(logger, fresh)
	},
}

func runMigrations(logger *zap.Logger, fresh bool) error {
	logger.Info("Starting database migrations")

	cfg := config.LoadConfig()
	logger.Info("Loaded the configuration", zap.Any("config", cfg))

	if fresh {
		logger.Warn("Fresh migration requested — dropping and recreating database!")

		if err := database.DropDatabase(cfg); err != nil {
			logger.Error("Failed to drop database", zap.Error(err))
			return fmt.Errorf("failed to drop database: %w", err)
		}

		if err := database.EnsureDatabaseExists(cfg); err != nil {
			logger.Error("Failed to recreate database", zap.Error(err))
			return fmt.Errorf("failed to recreate database: %w", err)
		}
	}

	db, err := database.ConnectToDatabase(cfg, true)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Terminal{}); err != nil {
		logger.Error("Failed to auto-migrate models", zap.Error(err))
		return fmt.Errorf("auto-migrate failed: %w", err)
	}
	logger.Info("Models migrated successfully")

	return nil
}
