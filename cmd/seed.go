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
	"math/rand"
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

	cfg := config.LoadConfig()
	logger.Info("Loaded the configuration", zap.Any("config", cfg))

	if err := database.EnsureDatabaseExists(cfg); err != nil {
		logger.Error("Database creation failed", zap.Error(err))
		return fmt.Errorf("failed to ensure database exists: %w", err)
	}

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
		{Username: "admin", Email: "admin@observv.me", Password: hashedPass, Role: "admin"},
		{Username: "user1", Email: "user1@observv.me", Password: hashedPass, Role: "user"},
		{Username: "user2", Email: "user2@observv.me", Password: hashedPass, Role: "user"},
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

	var terminalCount int64
	if err := db.Model(&models.Terminal{}).Count(&terminalCount).Error; err != nil {
		return fmt.Errorf("failed to check terminal count: %w", err)
	}
	if terminalCount > 0 {
		logger.Info("Terminals already exist, skipping", zap.Int64("count", terminalCount))
		return nil
	}

	var existingUsers []models.User
	if err := db.Find(&existingUsers).Error; err != nil {
		return fmt.Errorf("failed to fetch users for terminal assignment: %w", err)
	}
	if len(existingUsers) == 0 {
		logger.Warn("No users found; cannot assign terminals")
		return nil
	}

	totalTerminals := rand.Intn(11) + 5
	terminals := make([]models.Terminal, 0, totalTerminals)

	// One per user first
	for _, user := range existingUsers {
		terminals = append(terminals, models.Terminal{
			UserID:   user.ID,
			Status:   "offline",
			DataSent: 0,
		})
	}

	// Remaining distributed randomly
	remaining := totalTerminals - len(existingUsers)
	for i := 0; i < remaining; i++ {
		randomUser := existingUsers[rand.Intn(len(existingUsers))]
		terminals = append(terminals, models.Terminal{
			UserID:   randomUser.ID,
			Status:   "offline",
			DataSent: 0,
		})
	}

	if err := db.Create(&terminals).Error; err != nil {
		return fmt.Errorf("failed to insert terminals: %w", err)
	}

	logger.Info("Dummy terminals seeded", zap.Int("total", len(terminals)))
	return nil
}
