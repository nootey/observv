package main

import (
	"github.com/nootey/observv/pkg/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
)

var rootCmd = &cobra.Command{
	Use:     "observv",
	Short:   "A demo for the LGTM observability stack",
	Version: "1.0.0",
}

func init() {
	rootCmd.AddCommand(httpServerCmd)
	rootCmd.AddCommand(seedCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger, err := logging.NewLogger()
		if err != nil {
			log.Fatal("failed to initialize logger: %w", err)
		}
		defer logger.Sync()
		logger.Fatal("Failed to execute root command", zap.Error(err))
	}
}
