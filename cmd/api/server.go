package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type httpService struct {
	srv             *http.Server
	logger          *zap.Logger
	shutdownTimeout time.Duration
}

func (s *httpService) Name() string { return "http-server" }

func (s *httpService) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("server listening", zap.String("address", s.srv.Addr))
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()
		return s.srv.Shutdown(shutdownCtx)
	}
}
