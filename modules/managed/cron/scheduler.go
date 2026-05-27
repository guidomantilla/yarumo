package cron

import (
	"context"
	"sync"

	cron "github.com/robfig/cron/v3"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

// scheduler implements Scheduler. It embeds *cron.Cron by pointer so the
// underlying mutex and goroutine state are shared, not copied. The Start
// and Stop methods of cron.Cron are shadowed by lifecycle-aware versions;
// the remaining methods (AddFunc, AddJob, Schedule, Entries, ...) are
// promoted directly from the embedded *cron.Cron.
type scheduler struct {
	*cron.Cron

	name string
	done chan struct{}
	once sync.Once
}

// NewScheduler creates a new cron Scheduler with the given name and options.
func NewScheduler(name string, options ...cron.Option) Scheduler {
	cassert.NotEmpty(name, "name is empty")

	internal := cron.New(options...)

	return &scheduler{
		Cron: internal,
		name: name,
		done: make(chan struct{}),
	}
}

// Name returns the scheduler's identity used in logs.
func (s *scheduler) Name() string {
	cassert.NotNil(s, "scheduler is nil")

	return s.name
}

// Start begins the cron scheduler's internal goroutines and returns immediately.
// It satisfies the lifecycle.Component worker-style contract.
func (s *scheduler) Start(_ context.Context) error {
	cassert.NotNil(s, "scheduler is nil")

	s.Cron.Start()

	return nil
}

// Stop halts the cron scheduler and waits for running jobs to finish, bounded
// by ctx's deadline. It returns ErrShutdown wrapping ErrShutdownTimeout when
// ctx expires before the jobs drain.
func (s *scheduler) Stop(ctx context.Context) error {
	cassert.NotNil(s, "scheduler is nil")

	defer s.once.Do(func() { close(s.done) })

	stopCtx := s.Cron.Stop()
	select {
	case <-stopCtx.Done():
		return nil
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	}
}

// Done returns the channel that is closed after Stop has been called.
func (s *scheduler) Done() <-chan struct{} {
	cassert.NotNil(s, "scheduler is nil")

	return s.done
}
