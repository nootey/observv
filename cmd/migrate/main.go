package main

import (
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

	"github.com/nootey/observv/config"
	"github.com/nootey/observv/internal/store"
	"github.com/nootey/observv/pkg/logging"
)

const usage = "usage: migrate <up|down|status|reset|create NAME>"

func main() {
	if err := config.Load(); err != nil {
		panic("failed to load config: " + err.Error())
	}
	cfg := config.Get()

	logger := logging.InitLogger(cfg.App.Environment == "production", cfg.App.LogLevel)
	defer func(log *zap.Logger) {
		_ = log.Sync() // Ignore sync errors
	}(logger)
	logger = logger.Named("migrate")

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println(usage)
		os.Exit(1)
	}

	command := args[0]
	commandArgs := args[1:]

	db, err := store.NewDB(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database connection", zap.Error(err))
		}
	}()

	if err := goose.SetDialect("postgres"); err != nil {
		logger.Fatal("failed to set dialect", zap.Error(err))
	}

	switch command {
	case "up":
		if err := goose.Up(db, "migrations"); err != nil {
			logger.Fatal("migration up failed", zap.Error(err))
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if err := goose.Down(db, "migrations"); err != nil {
			logger.Fatal("migration down failed", zap.Error(err))
		}
		fmt.Println("Migration rolled back")

	case "status":
		if err := goose.Status(db, "migrations"); err != nil {
			logger.Fatal("migration status failed", zap.Error(err))
		}

	case "create":
		if len(commandArgs) == 0 {
			logger.Fatal("migration name required: go run ./cmd/migrate create <name>")
		}
		if err := goose.Create(db, "migrations", commandArgs[0], "sql"); err != nil {
			logger.Fatal("migration create failed", zap.Error(err))
		}
		fmt.Printf("Created migration: %s\n", commandArgs[0])

	case "reset":
		if err := goose.Reset(db, "migrations"); err != nil {
			logger.Fatal("migration reset failed", zap.Error(err))
		}
		fmt.Println("All migrations rolled back")

	default:
		fmt.Println(usage)
		os.Exit(1)
	}
}
