package managed

import (
	"context"
	"sync"

	ccron "github.com/guidomantilla/yarumo/cron"
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

func (c *cronWorker) Start(ctx context.Context) error {
	return c.c.Start(ctx)
}

func (c *cronWorker) Stop(ctx context.Context) error {
	defer c.once.Do(func() { close(c.done) })

	err := c.c.Stop(ctx)
	if err != nil {
		return ErrShutdown(ErrShutdownTimeout, ctx.Err())
	}

	return nil
}

func (c *cronWorker) Done() <-chan struct{} {
	return c.done
}
