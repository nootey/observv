package seeders

import (
	"context"
	"fmt"

	"github.com/nootey/observv/internal/database"

	"go.uber.org/zap"
)

func SeedTerminals(ctx context.Context, q *database.Queries, logger *zap.Logger, count int) error {
	logger.Info("seeding terminals", zap.Int("count", count))

	for i := 1; i <= count; i++ {
		if err := q.CreateTerminal(ctx, fmt.Sprintf("terminal-%02d", i)); err != nil {
			return err
		}
	}

	logger.Info("terminals seeded successfully")
	return nil
}
