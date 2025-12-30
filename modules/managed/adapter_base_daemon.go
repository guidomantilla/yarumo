package managed

import (
	"context"
	"sync"
)

type baseDaemon struct {
	done chan struct{}
	once sync.Once
}

func NewBaseDaemon() BaseDaemon {
	return &baseDaemon{
		done: make(chan struct{}),
	}
}

func (b *baseDaemon) Start() error {
	return nil
}

func (b *baseDaemon) Stop(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		b.once.Do(func() { close(b.done) })
		return nil
	}
}
func (b *baseDaemon) Done() <-chan struct{} {
	return b.done
}
