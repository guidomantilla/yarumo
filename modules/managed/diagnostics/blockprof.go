package diagnostics

import (
	"context"
	"runtime"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// blockprof is the canonical BlockProfiling implementation. It owns
// the runtime.SetBlockProfileRate toggle: Start enables sampling at
// the configured rate, Stop resets the rate to zero. Done closes when
// Stop has been called.
type blockprof struct {
	name string
	rate int
	done chan struct{}
	once sync.Once
}

// NewBlockProfiling creates a BlockProfiling with the given name and
// options. The name is used in logs and lifecycle events; it must be
// non-empty. The block-profile sampling rate is taken from
// WithBlockProfileRate (default: every blocking event, see
// blockProfileRateDefault).
func NewBlockProfiling(name string, options ...Option) BlockProfiling {
	cassert.NotEmpty(name, "name is empty")

	opts := NewOptions(options...)

	return &blockprof{
		name: name,
		rate: opts.blockProfileRate,
		done: make(chan struct{}),
	}
}

// Name returns the sampler's identity used in logs.
func (s *blockprof) Name() string {
	cassert.NotNil(s, "blockprof is nil")

	return s.name
}

// Start enables block-profile sampling at the configured rate and
// returns immediately. It satisfies the lifecycle.Component worker-
// style contract; Done is closed after Stop completes.
func (s *blockprof) Start(_ context.Context) error {
	cassert.NotNil(s, "blockprof is nil")

	runtime.SetBlockProfileRate(s.rate)

	return nil
}

// Stop disables block-profile sampling (rate=0) and closes Done. It is
// idempotent.
func (s *blockprof) Stop(_ context.Context) error {
	cassert.NotNil(s, "blockprof is nil")

	defer s.once.Do(func() { close(s.done) })

	runtime.SetBlockProfileRate(0)

	return nil
}

// Done returns the channel that is closed after Stop has been called.
func (s *blockprof) Done() <-chan struct{} {
	cassert.NotNil(s, "blockprof is nil")

	return s.done
}

// Rate returns the configured block-profile sampling rate.
func (s *blockprof) Rate() int {
	cassert.NotNil(s, "blockprof is nil")

	return s.rate
}
