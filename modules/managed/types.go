package managed

import (
	"context"
	"time"

	commoncron "github.com/guidomantilla/yarumo/common/cron"
	commongrpc "github.com/guidomantilla/yarumo/common/grpc"
	commonhttp "github.com/guidomantilla/yarumo/common/http"
)

type Worker interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Done() <-chan struct{}
}

type HttpServer interface {
	ListenAndServe(ctx context.Context) error
	ListenAndServeTLS(ctx context.Context, certFile string, keyFile string) error
	Stop(ctx context.Context) error
}

//

type BaseWorker interface {
	Worker
}

type CronWorker interface {
	Worker
}

type TraceFlightRecorderWorker interface {
	Worker
}

type GrpcServer interface {
	HttpServer
}

type ErrChan chan<- error

type StopFn func(ctx context.Context, timeout time.Duration)

type Component[T any] struct {
	name     string
	internal T
	metadata map[string]any
}

var (
	_ BuildFn[commongrpc.Server, GrpcServer]    = BuildGrpcServer
	_ BuildFn[commonhttp.Server, HttpServer]    = BuildHttpServer
	_ BuildFn[commoncron.Scheduler, CronWorker] = BuildCronWorker
	_ BuildFn[any, BaseWorker]                  = BuildBaseWorker
)

type BuildFn[I any, C any] func(ctx context.Context, name string, internal I, errChan ErrChan) (Component[C], StopFn, error)
