package runtime

import (
	"context"
	"fmt"
	"github.com/nootey/observv/internal/bootstrap"
	"github.com/nootey/observv/internal/http"
	"github.com/nootey/observv/pkg/config"
	"go.uber.org/zap"
	"os/signal"
	"syscall"
)

type HTTPServerRuntime struct {
	Logger *zap.Logger
	Config *config.Config
}

func NewHTTPServerRuntime(cfg *config.Config, logger *zap.Logger) *HTTPServerRuntime {
	return &HTTPServerRuntime{
		Config: cfg,
		Logger: logger,
	}
}

func (rt *HTTPServerRuntime) Run(context context.Context) error {
	ctx, stop := signal.NotifyContext(context, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	container := bootstrap.NewContainer(rt.Config)
	httpLogger := rt.Logger.Named("http").With(zap.String("component", "HTTP"))
	httpServer := http.NewServer(container, httpLogger)
	go httpServer.Start()

	<-ctx.Done()

	rt.Logger.Info("Interrupt signal received, shutting down HTTP server...")
	if err := httpServer.Shutdown(); err != nil {
		rt.Logger.Error("HTTP server shutdown failed", zap.Error(err))
		return fmt.Errorf("http server shutdown failed: %w", err)
	}
	rt.Logger.Info("HTTP server exiting")
	return nil
}
