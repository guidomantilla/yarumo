package managed

import (
	"context"
	"time"

	commoncron "github.com/guidomantilla/yarumo/common/cron"
	commongrpc "github.com/guidomantilla/yarumo/common/grpc"
	commonhttp "github.com/guidomantilla/yarumo/common/http"
)

type Daemon interface {
	Start() error
	Stop(ctx context.Context) error
	Done() <-chan struct{}
}

type HttpServer interface {
	ListenAndServe() error
	ListenAndServeTLS(certFile string, keyFile string) error
	Stop(ctx context.Context) error
}

//

type BaseDaemon interface {
	Daemon
}

type CronDaemon interface {
	Daemon
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
	_ BuildFn[commoncron.Scheduler, CronDaemon] = BuildCronServer
	_ BuildFn[any, BaseDaemon]                  = BuildBaseServer
)

type BuildFn[I any, C any] func(ctx context.Context, name string, internal I, errChan ErrChan) (Component[C], StopFn, error)
