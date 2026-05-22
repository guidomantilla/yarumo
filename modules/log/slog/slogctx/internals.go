package slogctx

import (
	"context"
	"log/slog"
	"sync"
)

// bagKeyType is an unexported context key type so external packages cannot
// collide with this package's value.
type bagKeyType struct{}

// bagKey is the singleton context key under which a *bag is stored.
//
//nolint:gochecknoglobals // unexported context key, idiomatic Go pattern.
var bagKey = bagKeyType{}

// bag is the thread-safe holder of slog.Attr values bound to a context. A bag
// is shared by all goroutines that share the same context — SetAttrs mutates
// it in place, while WithAttrs always produces a snapshot to keep parent
// contexts immutable.
type bag struct {
	mu    sync.RWMutex
	attrs []slog.Attr
}

// newBag returns an empty bag.
func newBag() *bag {
	return &bag{}
}

// snapshot returns a copy of the bag's attrs safe for concurrent readers.
func (b *bag) snapshot() []slog.Attr {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.attrs) == 0 {
		return nil
	}

	out := make([]slog.Attr, len(b.attrs))
	copy(out, b.attrs)

	return out
}

// append adds attrs to the bag.
func (b *bag) append(attrs ...slog.Attr) {
	if len(attrs) == 0 {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.attrs = append(b.attrs, attrs...)
}

// fromContext returns the bag carried by ctx, or nil when none is present.
func fromContext(ctx context.Context) *bag {
	if ctx == nil {
		return nil
	}

	value := ctx.Value(bagKey)
	if value == nil {
		return nil
	}

	carried, _ := value.(*bag)

	return carried
}
