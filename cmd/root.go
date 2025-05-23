package main

import (
	"context"
	"github.com/nootey/observv/pkg/config"
	"github.com/nootey/observv/pkg/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:     "observv",
	Short:   "A demo for the LGTM observability stack",
	Version: "1.0.0",
}

func init() {
	rootCmd.AddCommand(httpServerCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(seedCmd)
}

func Execute() {
	cfg := config.LoadConfig()

	logger := logging.NewLogger(cfg.Release)
	defer logger.Sync()
	rootCmd.SetContext(context.WithValue(context.Background(), "logger", logger))

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("Failed to execute root command", zap.Error(err))
	}
}
