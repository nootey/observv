package chaos_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nootey/observv/config"
	"github.com/nootey/observv/internal/apperr"
	"github.com/nootey/observv/internal/chaos"
)

func newHandler(t *testing.T, cfg config.ChaosConfig) *chaos.ChaosHandler {
	t.Helper()
	service, err := chaos.NewChaosService(cfg, zap.NewNop())
	require.NoError(t, err)
	return chaos.NewChaosHandler(service, apperr.NewResponder(zap.NewNop()))
}

func newRouter(t *testing.T, cfg config.ChaosConfig) *chi.Mux {
	t.Helper()
	r := chi.NewRouter()
	r.Mount("/chaos", newHandler(t, cfg).Routes())
	return r
}

func serve(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

// wrap puts a handler behind the chaos middleware. The flag is true once that handler ran.
func wrap(t *testing.T, cfg config.ChaosConfig) (http.Handler, *bool) {
	called := false
	return newHandler(t, cfg).Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})), &called
}

func getSettings(t *testing.T, r http.Handler) chaos.Settings {
	t.Helper()
	rec := serve(r, http.MethodGet, "/chaos/", "")
	require.Equal(t, http.StatusOK, rec.Code)

	var got chaos.Settings
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	return got
}

func TestNewChaosServiceRejectsInvalidConfig(t *testing.T) {
	cases := map[string]config.ChaosConfig{
		"error rate above 1": {ErrorRate: 2},
		"latency too high":   {LatencyMS: 60000},
	}

	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := chaos.NewChaosService(cfg, zap.NewNop())
			assert.Error(t, err)
		})
	}
}

func TestMiddlewarePassesThroughWhenOff(t *testing.T) {
	h, called := wrap(t, config.ChaosConfig{})

	rec := serve(h, http.MethodGet, "/", "")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, *called)
}

func TestMiddlewareInjectsErrors(t *testing.T) {
	h, called := wrap(t, config.ChaosConfig{ErrorRate: 1})

	rec := serve(h, http.MethodGet, "/", "")

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.False(t, *called)
}

func TestMiddlewareAddsLatency(t *testing.T) {
	h, called := wrap(t, config.ChaosConfig{LatencyMS: 30})

	start := time.Now()
	serve(h, http.MethodGet, "/", "")

	assert.GreaterOrEqual(t, time.Since(start), 30*time.Millisecond)
	assert.True(t, *called)
}

func TestUpdateSettings(t *testing.T) {
	r := newRouter(t, config.ChaosConfig{})

	rec := serve(r, http.MethodPost, "/chaos/", `{"latency_ms":200,"error_rate":0.5}`)
	require.Equal(t, http.StatusOK, rec.Code)

	assert.Equal(t, chaos.Settings{LatencyMS: 200, ErrorRate: 0.5}, getSettings(t, r))
}

func TestUpdateSettingsRejectsInvalid(t *testing.T) {
	cases := map[string]string{
		"error rate above 1": `{"latency_ms":0,"error_rate":1.5}`,
		"negative latency":   `{"latency_ms":-1,"error_rate":0}`,
		"latency too high":   `{"latency_ms":60000,"error_rate":0}`,
		"missing field":      `{"latency_ms":100}`,
		"unknown field":      `{"latencyMs":100,"error_rate":0}`,
		"not json":           `nope`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			r := newRouter(t, config.ChaosConfig{ErrorRate: 0.1})

			rec := serve(r, http.MethodPost, "/chaos/", body)
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			assert.Equal(t, 0.1, getSettings(t, r).ErrorRate, "a rejected update must keep the old settings")
		})
	}
}
