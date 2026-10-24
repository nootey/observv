package middleware

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

// RouteTelemetry sets the chi route on the otelhttp span and metrics. It must run inside the router.
func RouteTelemetry(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		pattern := routePattern(r)
		if pattern == "" {
			return
		}
		route := semconv.HTTPRoute(pattern)

		span := trace.SpanFromContext(r.Context())
		span.SetName(r.Method + " " + pattern)
		span.SetAttributes(route)

		if labeler, ok := otelhttp.LabelerFromContext(r.Context()); ok {
			labeler.Add(route)
		}
	})
}

// routePattern looks up the full route when a middleware answered before the subrouter routed.
func routePattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	pattern := rctx.RoutePattern()
	if !strings.HasSuffix(pattern, "/*") {
		return pattern
	}
	if found := rctx.Routes.Find(chi.NewRouteContext(), r.Method, r.URL.Path); found != "" {
		return found
	}
	return pattern
}
