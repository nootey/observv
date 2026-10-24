package simulator

import (
	"context"
	"errors"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/nootey/observv/config"
	"github.com/nootey/observv/internal/domains/terminal"
	"github.com/nootey/observv/pkg/logging"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("github.com/nootey/observv/internal/simulator")

var toggledStatus = map[string]string{
	terminal.StatusOnline:  terminal.StatusOffline,
	terminal.StatusOffline: terminal.StatusOnline,
}

type Terminal struct {
	id     int64
	name   string
	peers  []int64
	status string
	client *Client
	cfg    config.SimulatorConfig
	logger *zap.Logger
}

func NewFleet(terminals []terminal.Terminal, client *Client, cfg config.SimulatorConfig, logger *zap.Logger) []*Terminal {
	fleet := make([]*Terminal, 0, len(terminals))
	for _, t := range terminals {
		peers := make([]int64, 0, len(terminals)-1)
		for _, p := range terminals {
			if p.ID != t.ID {
				peers = append(peers, p.ID)
			}
		}

		fleet = append(fleet, &Terminal{
			id:     t.ID,
			name:   t.Name,
			peers:  peers,
			client: client,
			cfg:    cfg,
			logger: logger.With(zap.Int64("terminal_id", t.ID)),
		})
	}
	return fleet
}

func (t *Terminal) Name() string { return t.name }

func (t *Terminal) Run(ctx context.Context) error {
	interval := time.Duration(t.cfg.IntervalSeconds) * time.Second

	select {
	case <-ctx.Done():
		return nil
	case <-time.After(rand.N(interval)):
	}

	t.status = terminal.StatusOnline
	t.heartbeat(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.goOffline()
			return nil
		case <-ticker.C:
			t.tick(ctx)
		}
	}
}

func (t *Terminal) tick(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "terminal.tick", trace.WithAttributes(attribute.Int64("terminal.id", t.id)))
	defer span.End()

	if rand.Float64() < t.cfg.StatusFlipChance {
		t.status = toggledStatus[t.status]
	}

	t.heartbeat(ctx)

	if t.status == terminal.StatusOnline && len(t.peers) > 0 {
		t.transfer(ctx)
	}
}

func (t *Terminal) goOffline() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	t.status = terminal.StatusOffline
	t.heartbeat(ctx)
}

func (t *Terminal) heartbeat(ctx context.Context) {
	if err := t.client.Heartbeat(ctx, t.id, t.status); err != nil && ctx.Err() == nil {
		t.logger.Warn("heartbeat failed", zap.String("status", t.status), zap.Error(err), logging.Context(ctx))
	}
}

func (t *Terminal) transfer(ctx context.Context) {
	to := t.peers[rand.IntN(len(t.peers))]
	bytes := rand.Int64N(t.cfg.MaxTransferBytes) + 1

	err := t.client.Transfer(ctx, t.id, to, bytes)

	var apiErr *APIError
	switch {
	case err == nil:
	case errors.As(err, &apiErr) && apiErr.Status == http.StatusConflict:
		t.logger.Debug("transfer rejected", zap.Int64("to_terminal_id", to), zap.Error(err), logging.Context(ctx))
	case ctx.Err() == nil:
		t.logger.Warn("transfer failed", zap.Int64("to_terminal_id", to), zap.Error(err), logging.Context(ctx))
	}
}
