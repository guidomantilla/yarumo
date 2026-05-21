package lifecycle

import (
	"context"
	"sync"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// component is the basic Component implementation. Its Start is a no-op
// and Done closes when Stop is called. It serves as the building block
// for daemons that do not own a network listener.
type component struct {
	name string
	done chan struct{}
	once sync.Once
}

// NewComponent creates a new basic component with the given name.
func NewComponent(name string) Component {
	cassert.NotEmpty(name, "name is empty")

	return &component{
		name: name,
		done: make(chan struct{}),
	}
}

// Name returns the component's identity.
func (c *component) Name() string {
	cassert.NotNil(c, "component is nil")

	return c.name
}

// Start is a no-op for the basic component. It returns nil.
func (c *component) Start(_ context.Context) error {
	cassert.NotNil(c, "component is nil")

	return nil
}

// Stop closes the Done channel idempotently. It returns an ErrShutdown
// wrapping ErrShutdownTimeout when ctx has already expired on entry.
func (c *component) Stop(ctx context.Context) error {
	cassert.NotNil(c, "component is nil")

	defer c.once.Do(func() { close(c.done) })

	select {
	case <-ctx.Done():
		return ErrShutdown(ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (c *component) Done() <-chan struct{} {
	cassert.NotNil(c, "component is nil")

	return c.done
}
