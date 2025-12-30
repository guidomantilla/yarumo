package managed

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
)

type cronAdapter struct {
	c    *cron.Cron
	done chan struct{}
	once sync.Once
}

func NewCronDaemon(c *cron.Cron) CronDaemon {
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
