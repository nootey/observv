package terminal

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nootey/observv/internal/apperr"

	"github.com/go-chi/chi/v5"
)

var (
	errInvalidBody = apperr.New(apperr.KindInvalid, "invalid request body")
	errInvalidID   = apperr.New(apperr.KindInvalid, "invalid terminal id")
)

type TerminalHandler struct {
	service TerminalServiceInterface
	resp    *apperr.Responder
}

func NewTerminalHandler(srv TerminalServiceInterface, resp *apperr.Responder) *TerminalHandler {
	return &TerminalHandler{
		service: srv,
		resp:    resp,
	}
}

func (h *TerminalHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListTerminals)
	r.Get("/{id}", h.GetTerminal)
	r.Post("/{id}/heartbeat", h.RecordHeartbeat)
	r.Post("/{id}/transfers", h.Transfer)
	return r
}

func (h *TerminalHandler) ListTerminals(w http.ResponseWriter, r *http.Request) {
	terminals, err := h.service.ListTerminals(r.Context())
	if err != nil {
		h.resp.Error(w, r, err)
		return
	}

	h.resp.JSON(w, http.StatusOK, terminals)
}

func (h *TerminalHandler) GetTerminal(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.resp.Error(w, r, errInvalidID)
		return
	}

	terminal, err := h.service.GetTerminal(r.Context(), id)
	if err != nil {
		h.resp.Error(w, r, err)
		return
	}

	h.resp.JSON(w, http.StatusOK, terminal)
}

func (h *TerminalHandler) RecordHeartbeat(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.resp.Error(w, r, errInvalidID)
		return
	}

	var req HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.resp.Error(w, r, errInvalidBody)
		return
	}

	terminal, err := h.service.RecordHeartbeat(r.Context(), id, req.Status)
	if err != nil {
		h.resp.Error(w, r, err)
		return
	}

	h.resp.JSON(w, http.StatusOK, terminal)
}

func (h *TerminalHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.resp.Error(w, r, errInvalidID)
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.resp.Error(w, r, errInvalidBody)
		return
	}

	terminal, err := h.service.Transfer(r.Context(), id, req.ToID, req.Bytes)
	if err != nil {
		h.resp.Error(w, r, err)
		return
	}

	h.resp.JSON(w, http.StatusOK, terminal)
}
