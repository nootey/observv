package terminal

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/nootey/observv/internal/database"
)

const meterName = "github.com/nootey/observv/internal/domains/terminal"

const (
	attrTerminalID = attribute.Key("terminal.id")
	attrStatus     = attribute.Key("status")
)

type metrics struct {
	dataSent      metric.Int64Counter
	uptime        metric.Int64Counter
	statusChanges metric.Int64Counter
}

func newMetrics(provider metric.MeterProvider, queries *database.Queries) *metrics {
	meter := provider.Meter(meterName)
	m := &metrics{}
	var errs []error

	var err error
	m.dataSent, err = meter.Int64Counter("terminal.data.sent",
		metric.WithUnit("By"),
		metric.WithDescription("Bytes sent by a terminal."),
	)
	errs = append(errs, err)

	m.uptime, err = meter.Int64Counter("terminal.uptime",
		metric.WithUnit("s"),
		metric.WithDescription("Seconds a terminal was online."),
	)
	errs = append(errs, err)

	m.statusChanges, err = meter.Int64Counter("terminal.status.changes",
		metric.WithUnit("{change}"),
		metric.WithDescription("Status changes of a terminal. The status attribute is the new status."),
	)
	errs = append(errs, err)

	count, err := meter.Int64ObservableGauge("terminal.count",
		metric.WithUnit("{terminal}"),
		metric.WithDescription("Terminals by status, read from the database."),
	)
	errs = append(errs, err)

	_, err = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		rows, err := queries.CountTerminalsByStatus(ctx)
		if err != nil {
			return err
		}

		// Both statuses, so a status with no terminals reads 0.
		counts := map[string]int64{StatusOnline: 0, StatusOffline: 0}
		for _, row := range rows {
			counts[row.Status] = row.Count
		}
		for status, n := range counts {
			o.ObserveInt64(count, n, metric.WithAttributes(attrStatus.String(status)))
		}
		return nil
	}, count)
	errs = append(errs, err)

	if err := errors.Join(errs...); err != nil {
		otel.Handle(err)
	}
	return m
}

func (m *metrics) recordHeartbeat(ctx context.Context, prev, current *Terminal) {
	id := attrTerminalID.Int64(current.ID)

	if added := current.UptimeSeconds - prev.UptimeSeconds; added > 0 {
		m.uptime.Add(ctx, added, metric.WithAttributes(id))
	}
	if current.Status != prev.Status {
		m.statusChanges.Add(ctx, 1, metric.WithAttributes(id, attrStatus.String(current.Status)))
	}
}

func (m *metrics) recordTransfer(ctx context.Context, fromID, bytes int64) {
	m.dataSent.Add(ctx, bytes, metric.WithAttributes(attrTerminalID.Int64(fromID)))
}
