package seeders

import (
	"context"
	"database/sql"

	"github.com/nootey/observv/config"
	"github.com/nootey/observv/internal/database"

	"go.uber.org/zap"
)

func SeedDatabase(ctx context.Context, db *sql.DB, logger *zap.Logger, cfg *config.Config) error {
	queries := database.New(db)

	return SeedTerminals(ctx, queries, logger, cfg.Seed.Terminals)
}
