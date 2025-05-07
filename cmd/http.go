package main

import (
	"fmt"
	"github.com/nootey/observv/internal/runtime"
	"github.com/nootey/observv/pkg/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var httpServerCmd = &cobra.Command{
	Use:   "http",
	Short: "Run the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx := cmd.Context()

		logger, err := zap.NewProduction()
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}
		defer logger.Sync()

		cfg := config.LoadConfig()
		logger.Info("Configuration loaded",
			zap.String("port", cfg.HttpPort),
			zap.Bool("release", cfg.Release),
		)

		app := runtime.NewHTTPServerRuntime(cfg, logger)
		return app.Run(ctx)
	},
}
