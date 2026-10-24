package apperr

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestResolveKindStatus(t *testing.T) {
	cases := []struct {
		kind       Kind
		wantStatus int
	}{
		{KindInternal, http.StatusInternalServerError},
		{KindInvalid, http.StatusBadRequest},
		{KindValidation, http.StatusBadRequest},
		{KindUnauthorized, http.StatusUnauthorized},
		{KindForbidden, http.StatusForbidden},
		{KindNotFound, http.StatusNotFound},
		{KindConflict, http.StatusConflict},
	}

	for _, c := range cases {
		status, _ := Resolve(New(c.kind, "boom"))
		if status != c.wantStatus {
			t.Errorf("Resolve(kind=%d) status = %d, want %d", c.kind, status, c.wantStatus)
		}
	}
}

func TestResolveHidesInternalMessage(t *testing.T) {
	status, msg := Resolve(New(KindInternal, "database exploded at host db:5432"))
	if status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", status)
	}
	if msg != genericMessage {
		t.Errorf("msg = %q, want generic message %q", msg, genericMessage)
	}
}

func TestResolveKeepsClientMessage(t *testing.T) {
	_, msg := Resolve(New(KindInvalid, "email is required"))
	if msg != "email is required" {
		t.Errorf("msg = %q, want %q", msg, "email is required")
	}
}

func TestResolveUnknownError(t *testing.T) {
	status, msg := Resolve(errors.New("some raw error"))
	if status != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", status)
	}
	if msg != genericMessage {
		t.Errorf("msg = %q, want generic message %q", msg, genericMessage)
	}
}

func TestResolveWrappedAppError(t *testing.T) {
	base := New(KindNotFound, "user not found")
	wrapped := fmt.Errorf("loading profile: %w", base)

	status, msg := Resolve(wrapped)
	if status != http.StatusNotFound {
		t.Errorf("status = %d, want 404", status)
	}
	if msg != "user not found" {
		t.Errorf("msg = %q, want %q", msg, "user not found")
	}
}

func TestErrorUnwrapPreservesCause(t *testing.T) {
	cause := errors.New("pq: connection refused")
	err := Wrap(KindInternal, "failed to load user", cause)

	if !errors.Is(err, cause) {
		t.Error("errors.Is could not find the wrapped cause")
	}
}
