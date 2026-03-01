package managed

import (
	"context"
	"sync"

	cdiagnostics "github.com/guidomantilla/yarumo/common/diagnostics"
)

type traceFlightRecorderWorker struct {
	fr   cdiagnostics.TraceFlightRecorder
	done chan struct{}
	once sync.Once
}

// NewTraceFlightRecorderWorker creates a new managed trace flight recorder worker wrapping the given recorder.
func NewTraceFlightRecorderWorker(fr cdiagnostics.TraceFlightRecorder) TraceFlightRecorderWorker {
	return &traceFlightRecorderWorker{
		fr:   fr,
		done: make(chan struct{}),
	}
}

func (fr *traceFlightRecorderWorker) Start(_ context.Context) error {
	err := fr.fr.Start()
	if err != nil {
		return ErrStart(err)
	}
	return nil
}

func (fr *traceFlightRecorderWorker) Stop(ctx context.Context) error {
	defer fr.once.Do(func() { close(fr.done) })

	stopErrCh := make(chan error, 1)

	go func() {
		fr.fr.Stop()
		stopErrCh <- nil
	}()

	select {
	case <-ctx.Done():
		return ErrShutdown(ErrShutdownTimeout, ctx.Err())
	case err := <-stopErrCh:
		return err
	}
}

func (fr *traceFlightRecorderWorker) Done() <-chan struct{} {
	return fr.done
}
