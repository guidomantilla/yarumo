package servers

import (
	"net"
	"net/http"

	"github.com/qmdx00/lifecycle"
)

var (
	_ Server = (*cronServer)(nil)
	_ Server = (*grpcServer)(nil)
	_ Server = (*httpServer)(nil)
)

type Server interface {
	lifecycle.Server
}

type CronServer interface {
	Start()
	Stop()
}

type GrpcServer interface {
	Serve(lis net.Listener) error
	GracefulStop()
}

//

var (
	_ Application = (*lifecycle.App)(nil)
)

type Application interface {
	ID() string
	Name() string
	Version() string
	Metadata() map[string]string
	Attach(name string, server lifecycle.Server)
	Run() error
}

//

var (
	_ BuildBaseServerFn = BuildBaseServer
	_ BuildCronServerFn = BuildCronServer
	_ BuildHttpServerFn = BuildHttpServer
	_ BuildGrpcServerFn = BuildGrpcServer
)

type BuildBaseServerFn func() (string, Server)

type BuildCronServerFn func(cron CronServer) (string, Server)

type BuildHttpServerFn func(server *http.Server) (string, Server)

type BuildGrpcServerFn func(address string, server GrpcServer) (string, Server)
