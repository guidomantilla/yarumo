package managed

import (
	"context"
	"sync"

	commoncron "github.com/guidomantilla/yarumo/common/cron"
)

type cronWorker struct {
	c    commoncron.Scheduler
	done chan struct{}
	once sync.Once
}

func NewCronWorker(c commoncron.Scheduler) CronWorker {
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
	stopCtx := c.c.Stop()
	select {
	case <-stopCtx.Done():
		c.once.Do(func() { close(c.done) })
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *cronWorker) Done() <-chan struct{} {
	return c.done
}
