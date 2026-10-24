package worker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

type svc struct {
	name string
	run  func(ctx context.Context) error
}

func (s *svc) Name() string                  { return s.name }
func (s *svc) Run(ctx context.Context) error { return s.run(ctx) }

func TestSupervisorStopsAllOnContextCancel(t *testing.T) {
	var stopped int32

	blockUntilCancel := func(ctx context.Context) error {
		<-ctx.Done()
		atomic.AddInt32(&stopped, 1)
		return nil
	}

	sup := NewSupervisor(zap.NewNop())
	sup.Register(&svc{name: "a", run: blockUntilCancel})
	sup.Register(&svc{name: "b", run: blockUntilCancel})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	if err := sup.Run(ctx); err != nil {
		t.Fatalf("Run() = %v, want nil", err)
	}
	if got := atomic.LoadInt32(&stopped); got != 2 {
		t.Errorf("stopped services = %d, want 2", got)
	}
}

func TestSupervisorFirstErrorStopsGroup(t *testing.T) {
	wantErr := errors.New("boom")

	sup := NewSupervisor(zap.NewNop())
	sup.Register(&svc{name: "failing", run: func(ctx context.Context) error {
		return wantErr
	}})
	sup.Register(&svc{name: "blocking", run: func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	}})

	err := sup.Run(context.Background())
	if !errors.Is(err, wantErr) {
		t.Errorf("Run() = %v, want %v", err, wantErr)
	}
}

func TestSupervisorRecoversPanic(t *testing.T) {
	sup := NewSupervisor(zap.NewNop())
	sup.Register(&svc{name: "panicking", run: func(ctx context.Context) error {
		panic("kaboom")
	}})

	err := sup.Run(context.Background())
	if err == nil {
		t.Fatal("Run() = nil, want error from recovered panic")
	}
}
