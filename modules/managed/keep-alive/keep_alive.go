package keepalive

import (
	"context"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

// keepAlive is the basic lifecycle.Component implementation. Its Start is
// a no-op and Done closes when Stop is called. It serves as the building
// block for daemons that do not own a network listener.
type keepAlive struct {
	name string
	done chan struct{}
	once sync.Once
}

// NewKeepAlive creates a new basic lifecycle.Component with the given name.
func NewKeepAlive(name string) lifecycle.Component {
	cassert.NotEmpty(name, "name is empty")

	return &keepAlive{
		name: name,
		done: make(chan struct{}),
	}
}

// Name returns the component's identity.
func (k *keepAlive) Name() string {
	cassert.NotNil(k, "keepAlive is nil")

	return k.name
}

// Start is a no-op for the basic keep-alive component. It returns nil.
func (k *keepAlive) Start(_ context.Context) error {
	cassert.NotNil(k, "keepAlive is nil")

	return nil
}

// Stop closes the Done channel idempotently. It returns an ErrShutdown
// wrapping ErrShutdownTimeout when ctx has already expired on entry.
func (k *keepAlive) Stop(ctx context.Context) error {
	cassert.NotNil(k, "keepAlive is nil")

	defer k.once.Do(func() { close(k.done) })

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (k *keepAlive) Done() <-chan struct{} {
	cassert.NotNil(k, "keepAlive is nil")

	return k.done
}
