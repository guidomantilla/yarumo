// Package managed provides lifecycle management for application components.
package managed

import (
	"context"
	"time"

	ccron "github.com/guidomantilla/yarumo/common/cron"
	cdiagnostics "github.com/guidomantilla/yarumo/common/diagnostics"
	cgrpc "github.com/guidomantilla/yarumo/common/grpc"
	chttp "github.com/guidomantilla/yarumo/common/http"
)

// Worker defines the interface for a managed worker with start, stop, and done lifecycle methods.
type Worker interface {
	// Start begins the worker execution.
	Start(ctx context.Context) error
	// Stop gracefully stops the worker.
	Stop(ctx context.Context) error
	// Done returns a channel that is closed when the worker has stopped.
	Done() <-chan struct{}
}

// HttpServer defines the interface for a managed HTTP server.
type HttpServer interface {
	// ListenAndServe starts the HTTP server.
	ListenAndServe(ctx context.Context) error
	// ListenAndServeTLS starts the HTTP server with TLS.
	ListenAndServeTLS(ctx context.Context, certFile string, keyFile string) error
	// Stop gracefully stops the HTTP server.
	Stop(ctx context.Context) error
}

// BaseWorker defines the interface for a basic managed worker.
type BaseWorker interface {
	Worker
}

// CronWorker defines the interface for a cron-scheduled managed worker.
type CronWorker interface {
	Worker
}

// TraceFlightRecorderWorker defines the interface for a trace flight recorder managed worker.
type TraceFlightRecorderWorker interface {
	Worker
}

// GrpcServer defines the interface for a managed gRPC server.
type GrpcServer interface {
	HttpServer
}

// ErrChan is a send-only error channel used by builders to report startup errors.
type ErrChan chan<- error

// StopFn is the function type for stopping a managed component with a timeout.
type StopFn func(ctx context.Context, timeout time.Duration)

// Component holds a managed component with its name and internal value.
type Component[T any] struct {
	name     string
	internal T
}

// BuildFn is the function type for building a managed component from an internal dependency.
type BuildFn[I any, C any] func(ctx context.Context, name string, internal I, errChan ErrChan) (Component[C], StopFn, error)

var (
	_ BuildFn[cgrpc.Server, GrpcServer]                                    = BuildGrpcServer
	_ BuildFn[chttp.Server, HttpServer]                                    = BuildHttpServer
	_ BuildFn[ccron.Scheduler, CronWorker]                                 = BuildCronWorker
	_ BuildFn[any, BaseWorker]                                             = BuildBaseWorker
	_ BuildFn[cdiagnostics.TraceFlightRecorder, TraceFlightRecorderWorker] = BuildTraceFlightRecorderWorker
)

var (
	_ GrpcServer                = (*grpcAdapter)(nil)
	_ HttpServer                = (*httpAdapter)(nil)
	_ BaseWorker                = (*baseWorker)(nil)
	_ CronWorker                = (*cronWorker)(nil)
	_ TraceFlightRecorderWorker = (*traceFlightRecorderWorker)(nil)
)
