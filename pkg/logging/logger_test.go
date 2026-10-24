package logging

import (
	"context"
	"sync"
	"testing"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type memExporter struct {
	mu      sync.Mutex
	records []sdklog.Record
}

func (e *memExporter) Export(_ context.Context, records []sdklog.Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, r := range records {
		e.records = append(e.records, r.Clone())
	}
	return nil
}

func (e *memExporter) Shutdown(context.Context) error   { return nil }
func (e *memExporter) ForceFlush(context.Context) error { return nil }

func newTestLogger(level zapcore.Level) (*zap.Logger, *observer.ObservedLogs, *memExporter) {
	stdout, logs := observer.New(level)
	exporter := &memExporter{}
	provider := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewSimpleProcessor(exporter)))
	return WithOTel(zap.New(traceCore{stdout}), provider), logs, exporter
}

func spanContext(t *testing.T) (context.Context, trace.SpanContext) {
	t.Helper()
	traceID, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := trace.SpanIDFromHex("0102030405060708")
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled})
	return trace.ContextWithSpanContext(context.Background(), sc), sc
}

func TestWithOTelRecordCarriesTraceID(t *testing.T) {
	logger, _, exporter := newTestLogger(zapcore.InfoLevel)
	ctx, sc := spanContext(t)

	logger.Info("request completed", Context(ctx))

	if len(exporter.records) != 1 {
		t.Fatalf("got %d OTel records, want 1", len(exporter.records))
	}
	r := exporter.records[0]
	if r.TraceID() != sc.TraceID() {
		t.Errorf("record trace ID = %s, want %s", r.TraceID(), sc.TraceID())
	}
	if r.SpanID() != sc.SpanID() {
		t.Errorf("record span ID = %s, want %s", r.SpanID(), sc.SpanID())
	}
}

func TestStdoutGetsTraceIDsInsteadOfContext(t *testing.T) {
	logger, logs, _ := newTestLogger(zapcore.InfoLevel)
	ctx, sc := spanContext(t)

	logger.Info("request completed", Context(ctx))

	fields := logs.All()[0].ContextMap()
	if fields["trace_id"] != sc.TraceID().String() {
		t.Errorf("trace_id = %v, want %s", fields["trace_id"], sc.TraceID())
	}
	if fields["span_id"] != sc.SpanID().String() {
		t.Errorf("span_id = %v, want %s", fields["span_id"], sc.SpanID())
	}
	if _, ok := fields["context"]; ok {
		t.Error("stdout entry has a context field, want it replaced by trace_id and span_id")
	}
}

func TestWithOTelKeepsStdoutLevel(t *testing.T) {
	logger, logs, exporter := newTestLogger(zapcore.InfoLevel)

	logger.Debug("noise")
	logger.Info("signal")

	if logs.Len() != 1 {
		t.Errorf("got %d stdout entries, want 1", logs.Len())
	}
	if len(exporter.records) != 1 {
		t.Errorf("got %d OTel records, want 1: debug must not reach the bridge", len(exporter.records))
	}
}
