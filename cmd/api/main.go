package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"

	"github.com/nootey/observv/config"
	"github.com/nootey/observv/internal/apperr"
	"github.com/nootey/observv/internal/database"
	"github.com/nootey/observv/internal/domains/terminal"
	mWare "github.com/nootey/observv/internal/middleware"
	"github.com/nootey/observv/internal/store"
	"github.com/nootey/observv/internal/worker"
	"github.com/nootey/observv/pkg/logging"
)

type Handlers struct {
	Terminal *terminal.TerminalHandler
}

func main() {

	if err := config.Load(); err != nil {
		panic("failed to load config: " + err.Error())
	}
	cfg := config.Get()

	logger := logging.InitLogger(cfg.App.Environment == "production", cfg.App.LogLevel)
	defer func() {
		_ = logger.Sync()
	}()
	logger = logger.Named("api")

	logger.Info("starting application",
		zap.String("environment", cfg.App.Environment),
		zap.String("port", cfg.Server.Port),
	)

	db, err := store.NewDB(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database connection", zap.Error(err))
		}
	}()
	logger.Info("database connected")

	// Create SQLC queries instance
	queries := database.New(db)

	// Initialize services
	terminalService := terminal.NewTerminalService(queries)

	// Initialize handlers
	responder := apperr.NewResponder(logger)
	h := Handlers{
		Terminal: terminal.NewTerminalHandler(terminalService, responder),
	}

	r := setupRouter(&h, logger)

	srv := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	supervisor := worker.NewSupervisor(logger)
	supervisor.Register(&httpService{srv: srv, logger: logger, shutdownTimeout: 10 * time.Second})

	if err := supervisor.Run(ctx); err != nil {
		logger.Fatal("supervisor exited with error", zap.Error(err))
	}
	logger.Info("shutdown complete")
}

func setupRouter(h *Handlers, logger *zap.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Use(mWare.RequestLogger(logger.Named("http")))

		r.Mount("/terminals", h.Terminal.Routes())
	})

	logger.Info("router configured")
	return r
}
