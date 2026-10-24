package chaos

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"

	"github.com/nootey/observv/internal/apperr"
)

var (
	errInjected      = apperr.New(apperr.KindInternal, "chaos: injected failure")
	errInvalidBody   = apperr.New(apperr.KindInvalid, "invalid request body")
	errMissingFields = apperr.New(apperr.KindValidation, "latency_ms and error_rate are required")
)

type ChaosHandler struct {
	service *ChaosService
	resp    *apperr.Responder
}

func NewChaosHandler(srv *ChaosService, resp *apperr.Responder) *ChaosHandler {
	return &ChaosHandler{
		service: srv,
		resp:    resp,
	}
}

func (h *ChaosHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetSettings)
	r.Post("/", h.UpdateSettings)
	return r
}

func (h *ChaosHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.service.Delay(r.Context()) {
			return
		}

		if h.service.ShouldFail() {
			trace.SpanFromContext(r.Context()).AddEvent("chaos.error_injected")
			h.resp.Error(w, r, errInjected)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *ChaosHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	h.resp.JSON(w, http.StatusOK, h.service.Settings())
}

func (h *ChaosHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req UpdateSettingsRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		h.resp.Error(w, r, errInvalidBody)
		return
	}
	if req.LatencyMS == nil || req.ErrorRate == nil {
		h.resp.Error(w, r, errMissingFields)
		return
	}

	settings := Settings{LatencyMS: *req.LatencyMS, ErrorRate: *req.ErrorRate}
	if err := h.service.UpdateSettings(r.Context(), settings); err != nil {
		h.resp.Error(w, r, err)
		return
	}

	h.resp.JSON(w, http.StatusOK, settings)
}
