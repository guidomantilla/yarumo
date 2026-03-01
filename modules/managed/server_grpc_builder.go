package managed

import (
	"context"
	"time"

	cgrpc "github.com/guidomantilla/yarumo/common/grpc"
	clog "github.com/guidomantilla/yarumo/common/log"
)

// BuildGrpcServer creates a managed gRPC server component, starts it in a background goroutine, and returns a stop function.
func BuildGrpcServer(ctx context.Context, name string, internal cgrpc.Server, errChan ErrChan) (Component[GrpcServer], StopFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", name)

	grpcServer := Component[GrpcServer]{name: name, internal: NewGrpcServer(internal, "tcp")}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", name)
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", name)

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := grpcServer.internal.Stop(timeoutCtx)
		if err != nil {
			clog.Error(ctx, "shutdown failed", "stage", "shutdown", "component", name, "error", err)
		}
	}

	go func() {
		err := grpcServer.internal.ListenAndServe(ctx)
		if err != nil {
			clog.Error(ctx, "failed to listen or serve", "stage", "startup", "component", name, "error", err)
			select {
			case errChan <- err:
			default:
			}
		}
	}()

	return grpcServer, stopFn, nil
}
