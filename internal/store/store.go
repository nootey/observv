package store

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/XSAM/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/nootey/observv/config"

	_ "github.com/lib/pq"
)

func NewDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
	)

	attrs := otelsql.WithAttributes(
		semconv.DBSystemNamePostgreSQL,
		semconv.DBNamespace(cfg.DBName),
	)

	db, err := otelsql.Open("postgres", dsn,
		attrs,
		otelsql.WithSpanNameFormatter(queryName),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			OmitConnResetSession: true,
			OmitConnPrepare:      true,
			OmitRows:             true,
			SpanFilter:           inTrace,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if _, err := otelsql.RegisterDBStatsMetrics(db, attrs); err != nil {
		return nil, fmt.Errorf("failed to register db stats metrics: %w", err)
	}

	return db, nil
}

// queryName uses the sqlc query name ("-- name: GetTerminal :one") as the span name.
func queryName(_ context.Context, method otelsql.Method, query string) string {
	name, ok := strings.CutPrefix(query, "-- name: ")
	if !ok {
		return string(method)
	}
	name, _, _ = strings.Cut(name, " ")
	return name
}

// inTrace skips DB spans with no parent span, such as the terminal.count callback.
func inTrace(ctx context.Context, _ otelsql.Method, _ string, _ []driver.NamedValue) bool {
	return trace.SpanContextFromContext(ctx).IsValid()
}
