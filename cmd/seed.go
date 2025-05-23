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

		ctx := cmd.Context()
		logger := ctx.Value("logger").(*zap.Logger)

		return runSeeders(logger)
	},
}

func runSeeders(logger *zap.Logger) error {

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

	hashedPass, err := database.HashAndSaltPassword(cfg.RootUserPassword)
	if err != nil {
		logger.Error("Failed to hash root user password", zap.Error(err))
		return fmt.Errorf("failed to hash root user password: %w", err)
	}

	users := []models.User{
		{
			Username: "admin",
			Email:    "admin@observv.me",
			Password: hashedPass,
			Role:     "admin",
		},
		{
			Username: "user1",
			Email:    "user1@observv.me",
			Password: hashedPass,
			Role:     "user",
		},
		{
			Username: "user2",
			Email:    "user2@observv.me",
			Password: hashedPass,
			Role:     "user",
		},
	}

	for _, user := range users {
		var existing models.User
		err := db.Where("username = ? AND email = ?", user.Username, user.Email).First(&existing).Error
		if err == nil {
			logger.Info("User already exists, skipping", zap.String("username", user.Username))
			continue
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("Failed to query for existing user", zap.String("username", user.Username), zap.Error(err))
			return fmt.Errorf("failed to query user %s: %w", user.Username, err)
		}

		if err := db.Create(&user).Error; err != nil {
			logger.Error("Failed to create user", zap.String("username", user.Username), zap.Error(err))
			return fmt.Errorf("failed to create user %s: %w", user.Username, err)
		}
		logger.Info("Created user", zap.String("username", user.Username))
	}

	logger.Info("Database seeding complete.")

	return nil
}
