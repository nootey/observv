package worker

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type Service interface {
	Name() string
	Run(ctx context.Context) error
}

type Supervisor struct {
	logger   *zap.Logger
	services []Service
}

func NewSupervisor(logger *zap.Logger) *Supervisor {
	return &Supervisor{logger: logger}
}

func (s *Supervisor) Register(svc Service) {
	s.services = append(s.services, svc)
}

func (s *Supervisor) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg       sync.WaitGroup
		errOnce  sync.Once
		firstErr error
	)

	fail := func(err error) {
		if err != nil {
			errOnce.Do(func() { firstErr = err })
		}
		cancel()
	}

	for _, svc := range s.services {
		wg.Add(1)
		go func(svc Service) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("service panicked", zap.String("service", svc.Name()), zap.Any("panic", r))
					fail(fmt.Errorf("service %q panicked: %v", svc.Name(), r))
				}
			}()

			s.logger.Info("service started", zap.String("service", svc.Name()))
			err := svc.Run(ctx)
			if err != nil {
				s.logger.Error("service exited with error", zap.String("service", svc.Name()), zap.Error(err))
			} else {
				s.logger.Info("service stopped", zap.String("service", svc.Name()))
			}
			fail(err)
		}(svc)
	}

	wg.Wait()
	return firstErr
}
