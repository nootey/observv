package terminal

import (
	"context"
	"database/sql"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/nootey/observv/internal/apperr"
	"github.com/nootey/observv/internal/database"
)

var (
	errNotFound = apperr.New(apperr.KindNotFound, "terminal not found")
	errOffline  = apperr.New(apperr.KindConflict, "terminal is offline")
)

type TerminalServiceInterface interface {
	ListTerminals(ctx context.Context) ([]*Terminal, error)
	GetTerminal(ctx context.Context, id int64) (*Terminal, error)
	RecordHeartbeat(ctx context.Context, id int64, status string) (*Terminal, error)
	Transfer(ctx context.Context, fromID, toID, bytes int64) (*Terminal, error)
}

var _ TerminalServiceInterface = (*TerminalService)(nil)

type TerminalService struct {
	queries *database.Queries
	metrics *metrics
}

func NewTerminalService(queries *database.Queries) *TerminalService {
	return NewTerminalServiceWithMeter(queries, otel.GetMeterProvider())
}

func NewTerminalServiceWithMeter(queries *database.Queries, meterProvider metric.MeterProvider) *TerminalService {
	return &TerminalService{
		queries: queries,
		metrics: newMetrics(meterProvider, queries),
	}
}

func toTerminalModel(dbTerminal database.Terminal) *Terminal {
	t := &Terminal{
		ID:            dbTerminal.ID,
		Name:          dbTerminal.Name,
		Status:        dbTerminal.Status,
		DataSentBytes: dbTerminal.DataSentBytes,
		UptimeSeconds: dbTerminal.UptimeSeconds,
		CreatedAt:     dbTerminal.CreatedAt,
		UpdatedAt:     dbTerminal.UpdatedAt,
	}

	if dbTerminal.LastSeenAt.Valid {
		t.LastSeenAt = &dbTerminal.LastSeenAt.Time
	}

	return t
}

func (s *TerminalService) ListTerminals(ctx context.Context) ([]*Terminal, error) {
	dbTerminals, err := s.queries.ListTerminals(ctx)
	if err != nil {
		return nil, err
	}

	terminals := make([]*Terminal, len(dbTerminals))
	for i, dbTerminal := range dbTerminals {
		terminals[i] = toTerminalModel(dbTerminal)
	}

	return terminals, nil
}

func (s *TerminalService) GetTerminal(ctx context.Context, id int64) (*Terminal, error) {
	dbTerminal, err := s.queries.GetTerminal(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
	}
	if err != nil {
		return nil, err
	}

	return toTerminalModel(dbTerminal), nil
}

func (s *TerminalService) RecordHeartbeat(ctx context.Context, id int64, status string) (*Terminal, error) {
	if status != StatusOnline && status != StatusOffline {
		return nil, apperr.New(apperr.KindValidation, "status must be online or offline")
	}

	prev, err := s.GetTerminal(ctx, id)
	if err != nil {
		return nil, err
	}

	updated, err := s.queries.RecordHeartbeat(ctx, database.RecordHeartbeatParams{
		ID:     id,
		Status: status,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
	}
	if err != nil {
		return nil, err
	}

	current := toTerminalModel(updated)
	s.metrics.recordHeartbeat(ctx, prev, current)

	return current, nil
}

func (s *TerminalService) Transfer(ctx context.Context, fromID, toID, bytes int64) (*Terminal, error) {
	if bytes <= 0 {
		return nil, apperr.New(apperr.KindValidation, "bytes must be greater than zero")
	}
	if fromID == toID {
		return nil, apperr.New(apperr.KindInvalid, "a terminal cannot transfer to itself")
	}

	from, err := s.GetTerminal(ctx, fromID)
	if err != nil {
		return nil, err
	}
	to, err := s.GetTerminal(ctx, toID)
	if err != nil {
		return nil, err
	}

	if from.Status != StatusOnline || to.Status != StatusOnline {
		return nil, errOffline
	}

	updated, err := s.queries.AddDataSent(ctx, database.AddDataSentParams{ID: fromID, Bytes: bytes})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
	}
	if err != nil {
		return nil, err
	}

	s.metrics.recordTransfer(ctx, fromID, bytes)

	return toTerminalModel(updated), nil
}
