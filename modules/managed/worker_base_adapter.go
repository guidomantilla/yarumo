package managed

import (
	"context"
	"sync"
)

type baseWorker struct {
	done chan struct{}
	once sync.Once
}

// NewBaseWorker creates a new managed base worker.
func NewBaseWorker() BaseWorker {
	return &baseWorker{
		done: make(chan struct{}),
	}
}

func (b *baseWorker) Start(_ context.Context) error {
	return nil
}

func (b *baseWorker) Stop(ctx context.Context) error {
	defer b.once.Do(func() { close(b.done) })

	select {
	case <-ctx.Done():
		return ErrShutdown(ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

func (b *baseWorker) Done() <-chan struct{} {
	return b.done
}
