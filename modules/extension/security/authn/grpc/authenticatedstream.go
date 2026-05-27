package grpc

import (
	"context"

	"google.golang.org/grpc"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// authenticatedStream wraps a grpc.ServerStream so its Context()
// returns the principal-bearing ctx. Every other method is delegated to
// the embedded ServerStream.
type authenticatedStream struct {
	grpc.ServerStream

	ctx context.Context
}

// Context returns the per-RPC ctx augmented with the validated
// *Principal.
func (s *authenticatedStream) Context() context.Context {
	cassert.NotNil(s, "authenticatedStream is nil")

	return s.ctx
}
