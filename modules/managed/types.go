package managed

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"net/http"
	_ "net/http/pprof"
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
	RegisterService(desc *grpc.ServiceDesc, impl any)
}

type ErrChan chan<- error

type StopFn func(ctx context.Context, timeout time.Duration)

type Component[T any] struct {
	name     string
	internal T
	metadata map[string]any
}
