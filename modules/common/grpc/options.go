package grpc

import (
	"google.golang.org/grpc"
)

type serviceRegistration struct {
	service    any
	descriptor *grpc.ServiceDesc
}

// Option is a functional option for configuring gRPC Options.
type Option func(opts *Options)

// Options holds the configuration for creating a gRPC server.
type Options struct {
	services      []serviceRegistration
	serverOptions []grpc.ServerOption
}

// NewOptions creates a new Options applying all provided Option functions.
func NewOptions(opts ...Option) *Options {
	o := &Options{
		services:      make([]serviceRegistration, 0),
		serverOptions: make([]grpc.ServerOption, 0),
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// WithService returns an Option that registers a service with its descriptor.
// If service or descriptor is nil the option is a no-op.
func WithService(service any, descriptor *grpc.ServiceDesc) Option {
	return func(opts *Options) {
		if service == nil || descriptor == nil {
			return
		}

		opts.services = append(opts.services, serviceRegistration{
			service:    service,
			descriptor: descriptor,
		})
	}
}

// WithServerOption returns an Option that appends gRPC server options.
// Nil individual options are silently ignored.
func WithServerOption(serverOpts ...grpc.ServerOption) Option {
	return func(opts *Options) {
		for _, so := range serverOpts {
			if so == nil {
				continue
			}

			opts.serverOptions = append(opts.serverOptions, so)
		}
	}
}
