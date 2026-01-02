package managed

import (
	"context"
	"fmt"
	"sync"
)

type baseWorker struct {
	done chan struct{}
	once sync.Once
}

func NewBaseWorker() BaseWorker {
	return &baseWorker{
		done: make(chan struct{}),
	}
}

func (b *baseWorker) Start(_ context.Context) error {
	return nil
}

func (b *baseWorker) Stop(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	default:
		b.once.Do(func() { close(b.done) })
		return nil
	}
}
func (b *baseWorker) Done() <-chan struct{} {
	return b.done
}
