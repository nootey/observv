package apperr

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type ErrorResponse struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Title   string `json:"title,omitempty"`
	Message string `json:"message,omitempty"`
}

type Responder struct {
	logger *zap.Logger
}

func NewResponder(logger *zap.Logger) *Responder {
	return &Responder{logger: logger}
}

func (rp *Responder) JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func (rp *Responder) Success(w http.ResponseWriter, status int, message string) {
	rp.JSON(w, status, SuccessResponse{Title: "Success", Message: message})
}

func (rp *Responder) Error(w http.ResponseWriter, r *http.Request, err error) {
	status, message := Resolve(err)
	if status >= http.StatusInternalServerError {
		rp.logger.Error("request failed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("request_id", middleware.GetReqID(r.Context())),
			zap.Error(err),
		)
	}
	rp.JSON(w, status, ErrorResponse{Title: "Error", Message: message})
}
