package main

import (
	"github.com/nootey/observv/internal/runtime"
	"github.com/nootey/observv/pkg/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Run the HTTP server and simulate terminal data transfer",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		logger := ctx.Value("logger").(*zap.Logger)

		cfg := config.LoadConfig()
		logger.Info("Configuration loaded for demo",
			zap.String("port", cfg.HttpPort),
			zap.Bool("release", cfg.Release),
		)

		app := runtime.NewHTTPServerRuntime(cfg, logger)
		return app.Run(ctx, true)
	},
}
