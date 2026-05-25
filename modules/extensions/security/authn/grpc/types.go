package grpc

import (
	"google.golang.org/grpc"

	"github.com/guidomantilla/yarumo/security/authn"
)

var (
	_ UnaryInterceptorFactoryFn  = NewUnaryInterceptor
	_ StreamInterceptorFactoryFn = NewStreamInterceptor
)

// UnaryInterceptorFactoryFn is the function type for
// NewUnaryInterceptor.
type UnaryInterceptorFactoryFn func(authenticator authn.Authenticator, options ...Option) grpc.UnaryServerInterceptor

// StreamInterceptorFactoryFn is the function type for
// NewStreamInterceptor.
type StreamInterceptorFactoryFn func(authenticator authn.Authenticator, options ...Option) grpc.StreamServerInterceptor
