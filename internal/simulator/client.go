package simulator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nootey/observv/internal/apperr"
	"github.com/nootey/observv/internal/domains/terminal"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

// APIError is a non-2xx answer from the API.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("api returned %d: %s", e.Status, e.Message)
}

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout:   10 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

func (c *Client) ListTerminals(ctx context.Context) ([]terminal.Terminal, error) {
	var terminals []terminal.Terminal
	err := c.do(ctx, http.MethodGet, "/api/terminals", nil, &terminals)
	return terminals, err
}

func (c *Client) Heartbeat(ctx context.Context, id int64, status string) error {
	path := fmt.Sprintf("/api/terminals/%d/heartbeat", id)
	return c.do(ctx, http.MethodPost, path, terminal.HeartbeatRequest{Status: status}, nil)
}

func (c *Client) Transfer(ctx context.Context, fromID, toID, bytes int64) error {
	path := fmt.Sprintf("/api/terminals/%d/transfers", fromID)
	return c.do(ctx, http.MethodPost, path, terminal.TransferRequest{ToID: toID, Bytes: bytes}, nil)
}

func (c *Client) WaitForTerminals(ctx context.Context, logger *zap.Logger) ([]terminal.Terminal, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		terminals, err := c.ListTerminals(ctx)
		if err == nil {
			return terminals, nil
		}
		logger.Info("waiting for api", zap.String("url", c.baseURL), zap.Error(err))

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusMultipleChoices {
		var errResp apperr.ErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return &APIError{Status: resp.StatusCode, Message: errResp.Message}
	}

	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
