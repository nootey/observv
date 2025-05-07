package main

import (
	"errors"
	"fmt"
	"github.com/nootey/observv/internal/models"
	"github.com/nootey/observv/pkg/config"
	"github.com/nootey/observv/pkg/database"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Run database seeders",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSeeders()
	},
}

func runSeeders() error {

	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("Starting database seeding")

	// Load Configuration
	cfg := config.LoadConfig()
	logger.Info("Loaded the configuration", zap.Any("config", cfg))

	// Ensure database exists
	if err := database.EnsureDatabaseExists(cfg); err != nil {
		logger.Error("Database creation failed", zap.Error(err))
		return fmt.Errorf("failed to ensure database exists: %w", err)
	}

	// Connect to MySQL using GORM
	db, err := database.ConnectToDatabase(cfg, true)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %v", err)
	}

	//Migrate the tables
	if err := db.AutoMigrate(&models.User{}); err != nil {
		logger.Error("Failed to auto-migrate User model", zap.Error(err))
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}

	// Check if the admin user already exists
	var existing models.User
	err = db.Where("username = ? AND email = ?", "admin", "admin@observv.me").First(&existing).Error
	if err == nil {
		logger.Info("Admin user already exists, skipping seeding")
		return nil
	}
	if err != nil && err.Error() != "record not found" && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("Failed to query for existing admin user", zap.Error(err))
		return fmt.Errorf("failed to query admin user: %w", err)
	}

	hashedPass, err := database.HashAndSaltPassword(cfg.RootUserPassword)
	if err != nil {
		logger.Error("Failed to hash root user password", zap.Error(err))
		return fmt.Errorf("failed to hash root user password: %w", err)
	}

	admin := models.User{
		Username: "admin",
		Password: hashedPass,
		Email:    "admin@observv.me",
		Role:     "admin",
	}
	if err := db.Create(&admin).Error; err != nil {
		logger.Error("Failed to create default admin user", zap.Error(err))
		return fmt.Errorf("failed to create default admin user: %w", err)
	}

	logger.Info("Database seeding complete.")

	return nil
}
