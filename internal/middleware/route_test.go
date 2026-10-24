package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"

	"github.com/nootey/observv/internal/middleware"
)

type instrumented struct {
	handler http.Handler
	spans   *tracetest.SpanRecorder
	metrics *sdkmetric.ManualReader
}

// rejectAll answers before routing, as chaos does when it injects an error.
func rejectAll(http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
}

// newInstrumented mirrors cmd/api: otelhttp outside the router, RouteTelemetry inside it.
func newInstrumented() instrumented {
	spans := tracetest.NewSpanRecorder()
	metrics := sdkmetric.NewManualReader()

	terminals := chi.NewRouter()
	terminals.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {})

	rejected := chi.NewRouter()
	rejected.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {})

	r := chi.NewRouter()
	r.Use(middleware.RouteTelemetry)
	r.Mount("/terminals", terminals)
	r.With(rejectAll).Mount("/rejected", rejected)

	return instrumented{
		handler: otelhttp.NewHandler(r, "test",
			otelhttp.WithTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans))),
			otelhttp.WithMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(metrics))),
		),
		spans:   spans,
		metrics: metrics,
	}
}

func (i instrumented) get(path string) {
	i.handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
}

func (i instrumented) durationPoint(t *testing.T) metricdata.HistogramDataPoint[float64] {
	t.Helper()
	var rm metricdata.ResourceMetrics
	require.NoError(t, i.metrics.Collect(context.Background(), &rm))
	for _, scope := range rm.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name == "http.server.request.duration" {
				points := m.Data.(metricdata.Histogram[float64]).DataPoints
				require.Len(t, points, 1)
				return points[0]
			}
		}
	}
	t.Fatal("http.server.request.duration not recorded")
	return metricdata.HistogramDataPoint[float64]{}
}

func TestRouteTelemetryUsesRoutePattern(t *testing.T) {
	inst := newInstrumented()

	inst.get("/terminals/7")

	ended := inst.spans.Ended()
	require.Len(t, ended, 1)
	assert.Equal(t, "GET /terminals/{id}", ended[0].Name())
	assert.Contains(t, ended[0].Attributes(), semconv.HTTPRoute("/terminals/{id}"))

	attrs := inst.durationPoint(t).Attributes
	route, ok := attrs.Value(semconv.HTTPRouteKey)
	require.True(t, ok, "duration metric has no http.route")
	assert.Equal(t, "/terminals/{id}", route.AsString())
}

func TestRouteTelemetryFindsRouteAfterEarlyAnswer(t *testing.T) {
	inst := newInstrumented()

	inst.get("/rejected/7")

	ended := inst.spans.Ended()
	require.Len(t, ended, 1)
	assert.Equal(t, "GET /rejected/{id}", ended[0].Name())

	attrs := inst.durationPoint(t).Attributes
	route, ok := attrs.Value(semconv.HTTPRouteKey)
	require.True(t, ok, "duration metric has no http.route")
	assert.Equal(t, "/rejected/{id}", route.AsString())
}

func TestRouteTelemetryUnmatchedRoute(t *testing.T) {
	inst := newInstrumented()

	inst.get("/nope")

	ended := inst.spans.Ended()
	require.Len(t, ended, 1)
	assert.Equal(t, "GET", ended[0].Name())

	attrs := inst.durationPoint(t).Attributes
	_, ok := attrs.Value(semconv.HTTPRouteKey)
	assert.False(t, ok, "unmatched request must not get an http.route")
}
