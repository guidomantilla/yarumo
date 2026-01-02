package managed

import (
	"context"
	"fmt"
	"sync"

	"github.com/guidomantilla/yarumo/common/diagnostics"
)

type traceFlightRecorderWorker struct {
	fr   diagnostics.TraceFlightRecorder
	done chan struct{}
	once sync.Once
}

func NewTraceFlightRecorderWorker(fr diagnostics.TraceFlightRecorder) TraceFlightRecorderWorker {
	return &traceFlightRecorderWorker{
		fr:   fr,
		done: make(chan struct{}),
	}
}

func (fr *traceFlightRecorderWorker) Start(_ context.Context) error {
	err := fr.fr.Start()
	if err != nil {
		return fmt.Errorf("failed to start trace flight recorder: %w", err)
	}
	return nil
}

func (fr *traceFlightRecorderWorker) Stop(ctx context.Context) error {
	stopErrCh := make(chan error, 1)

	go func() {
		fr.fr.Stop()
		stopErrCh <- nil
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	case err := <-stopErrCh:
		fr.once.Do(func() { close(fr.done) })
		return err
	}
}

func (fr *traceFlightRecorderWorker) Done() <-chan struct{} {
	return fr.done
}
