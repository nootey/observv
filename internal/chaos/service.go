package chaos

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/nootey/observv/config"
	"github.com/nootey/observv/internal/apperr"
	"github.com/nootey/observv/pkg/logging"
)

// Below the server write timeout.
const maxLatencyMS = 5000

var tracer = otel.Tracer("github.com/nootey/observv/internal/chaos")

type ChaosService struct {
	settings atomic.Pointer[Settings]
	logger   *zap.Logger
}

func NewChaosService(cfg config.ChaosConfig, logger *zap.Logger) (*ChaosService, error) {
	initial := Settings{LatencyMS: cfg.LatencyMS, ErrorRate: cfg.ErrorRate}
	if err := validateSettings(initial); err != nil {
		return nil, err
	}

	s := &ChaosService{logger: logger}
	s.settings.Store(&initial)
	return s, nil
}

func (s *ChaosService) Settings() Settings {
	return *s.settings.Load()
}

func (s *ChaosService) UpdateSettings(ctx context.Context, settings Settings) error {
	if err := validateSettings(settings); err != nil {
		return err
	}

	s.settings.Store(&settings)
	s.logger.Info("chaos settings changed",
		zap.Int("latency_ms", settings.LatencyMS),
		zap.Float64("error_rate", settings.ErrorRate),
		logging.Context(ctx),
	)
	return nil
}

// Delay waits for the set latency in a chaos.latency span. It returns false
// if ctx ends first.
func (s *ChaosService) Delay(ctx context.Context) bool {
	latency := time.Duration(s.Settings().LatencyMS) * time.Millisecond
	if latency == 0 {
		return true
	}

	ctx, span := tracer.Start(ctx, "chaos.latency",
		trace.WithAttributes(attribute.Int64("chaos.latency_ms", latency.Milliseconds())),
	)
	defer span.End()

	timer := time.NewTimer(latency)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func (s *ChaosService) ShouldFail() bool {
	rate := s.Settings().ErrorRate
	return rate > 0 && rand.Float64() < rate
}

func validateSettings(s Settings) error {
	if s.LatencyMS < 0 || s.LatencyMS > maxLatencyMS {
		return apperr.New(apperr.KindValidation, fmt.Sprintf("latency_ms must be between 0 and %d", maxLatencyMS))
	}
	if s.ErrorRate < 0 || s.ErrorRate > 1 {
		return apperr.New(apperr.KindValidation, "error_rate must be between 0 and 1")
	}
	return nil
}
