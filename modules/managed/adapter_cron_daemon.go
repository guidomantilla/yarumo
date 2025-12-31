package managed

import (
	"context"
	"sync"

	commoncron "github.com/guidomantilla/yarumo/common/cron"
)

type cronAdapter struct {
	c    commoncron.Scheduler
	done chan struct{}
	once sync.Once
}

func NewCronDaemon(c commoncron.Scheduler) CronDaemon {
	return &cronAdapter{
		c:    c,
		done: make(chan struct{}),
	}
}

func (c *cronAdapter) Start() error {
	c.c.Start()
	return nil
}

func (c *cronAdapter) Stop(ctx context.Context) error {
	stopCtx := c.c.Stop()
	select {
	case <-stopCtx.Done():
		c.once.Do(func() { close(c.done) })
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *cronAdapter) Done() <-chan struct{} {
	return c.done
}
