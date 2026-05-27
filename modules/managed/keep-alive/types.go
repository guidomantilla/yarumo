// Package keepalive provides a basic lifecycle.Component whose Start is a
// no-op and whose Done channel closes when Stop is called. It serves as the
// building block for daemons that do not own a network listener (heartbeats,
// long-running workers, application keep-alive loops).
//
// The component contract — Component, Start, Stop, Done, the ErrShutdown
// family and the lifecycle.Build helper — lives in
// `modules/common/lifecycle`. This module is a thin wrapper that returns a
// ready-made `lifecycle.Component` implementation; it owns no interface of
// its own.
package keepalive

import (
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

var (
	_ lifecycle.Component = (*keepAlive)(nil)
)
