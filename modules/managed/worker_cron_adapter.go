package managed

import (
	"context"
	"sync"

	ccron "github.com/guidomantilla/yarumo/common/cron"
)

type cronWorker struct {
	c    ccron.Scheduler
	done chan struct{}
	once sync.Once
}

// NewCronWorker creates a new managed cron worker wrapping the given scheduler.
func NewCronWorker(c ccron.Scheduler) CronWorker {
	return &cronWorker{
		c:    c,
		done: make(chan struct{}),
	}
}

func (c *cronWorker) Start(_ context.Context) error {
	c.c.Start()
	return nil
}

func (c *cronWorker) Stop(ctx context.Context) error {
	defer c.once.Do(func() { close(c.done) })

	stopCtx := c.c.Stop()
	select {
	case <-stopCtx.Done():
		return nil
	case <-ctx.Done():
		return ErrShutdown(ErrShutdownTimeout, ctx.Err())
	}
}

func (c *cronWorker) Done() <-chan struct{} {
	return c.done
}
