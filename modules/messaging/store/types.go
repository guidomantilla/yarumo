// Package store provides support-primitive storage contracts for the
// messaging EIP patterns and a canonical in-memory backend for each
// contract.
//
// Two interfaces live here:
//
//   - MessageStore[T] persists full Message[T] envelopes keyed by an
//     opaque string identifier. It is the storage primitive the Claim
//     Check EIP pattern (YA-0229) uses to offload large payloads from
//     the in-process channel: the producer stores the heavy message
//     and forwards a reference (the key); downstream consumers
//     retrieve the original when they actually need the payload.
//   - MetadataStore records boolean presence of a key with TTL. It is
//     the storage primitive the Idempotent Receiver EIP pattern
//     (YA-0228) uses for MessageID dedup history: the receiver checks
//     whether it has seen a message before, and if not, records it
//     with a TTL so the dedup window is bounded.
//
// # In-memory canonical implementations
//
// The package ships an in-memory backend for each interface:
//
//   - NewInMemoryMessageStore[T] returns a MessageStore[T] backed by a
//     mutex-guarded map. It is purely passive — no goroutines, no
//     lifecycle.
//   - NewInMemoryMetadataStore returns a MetadataStore backed by a
//     mutex-guarded map plus a background sweeper goroutine that
//     evicts expired keys at the configured interval. Because of the
//     sweeper, the in-memory metadata store implements
//     common/lifecycle.Component (worker-style): Start spawns the
//     sweeper; Stop drains it.
//
// Heavy-dep backends (Redis, Postgres, S3, …) belong in
// extension/messaging/store/<backend>/ — they get their own go-module
// per the workspace MVS isolation rule. This sub-package owns no
// external dependencies beyond core/common.
//
// # Concurrency
//
// All public types in this package are safe for concurrent use by
// multiple goroutines.
package store

import (
	"context"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// Type compliance: the in-memory backends satisfy their respective
// interfaces; the metadata backend additionally satisfies
// lifecycle.Component because of the sweeper goroutine; the free
// error factories match their declared Fn aliases.
var (
	_ MessageStore[any]   = (*inMemoryMessageStore[any])(nil)
	_ MetadataStore       = (*inMemoryMetadataStore)(nil)
	_ lifecycle.Component = (*inMemoryMetadataStore)(nil)

	_ ErrStoreFn    = ErrStore
	_ ErrNotFoundFn = ErrNotFound
)

// MessageStore persists full Message[T] envelopes keyed by an opaque
// string identifier. Implementations must be safe for concurrent use
// by multiple goroutines.
//
// MessageStore is the storage primitive consumed by the Claim Check
// EIP pattern: the producer Put-s the heavy message and forwards only
// the key over the in-process channel; downstream consumers Get the
// original envelope when they actually need the payload, and
// Delete-it once it is no longer needed.
//
// Implementations may store envelopes in process memory, in a
// distributed cache (Redis), in a relational database, or in object
// storage — the contract makes no commitment about durability or
// replication. Consumers that need stronger guarantees pick the
// backend accordingly.
type MessageStore[T any] interface {
	// Put stores msg under key. It overwrites any value previously
	// stored at the same key. ctx gates the operation for backends
	// that perform I/O; the in-memory backend ignores ctx beyond
	// honoring its cancellation if it expires before the call returns.
	Put(ctx context.Context, key string, msg messaging.Message[T]) error
	// Get returns the message previously stored at key. It returns an
	// error wrapping ErrStoreNotFound when the key does not exist,
	// matchable via errors.Is(err, ErrStoreNotFound).
	Get(ctx context.Context, key string) (messaging.Message[T], error)
	// Delete removes the message stored at key. Delete is idempotent:
	// it returns nil whether or not the key was present before the
	// call.
	Delete(ctx context.Context, key string) error
}

// MetadataStore records boolean presence of a key with TTL.
// Implementations must be safe for concurrent use by multiple
// goroutines.
//
// MetadataStore is the storage primitive consumed by the Idempotent
// Receiver EIP pattern for MessageID dedup history. The store is
// intentionally minimal — it does not store payloads, only "have I
// seen this key before, and is the record still fresh?". Keys are
// evicted after their TTL expires.
//
// Implementations may store the presence set in process memory, in a
// distributed cache (Redis SETEX), or in any backend that supports
// per-key TTL semantics. Consumers that need a longer or unbounded
// dedup window configure a larger TTL or pick a durable backend.
type MetadataStore interface {
	// Has reports whether key is currently present in the store. A
	// key that was Add-ed but whose TTL has expired returns false.
	// ctx gates the operation for backends that perform I/O; the
	// in-memory backend ignores ctx beyond honoring its cancellation
	// if it expires before the call returns.
	Has(ctx context.Context, key string) (bool, error)
	// Add records key with the given TTL. If the key already exists,
	// the TTL is refreshed (last writer wins). ttl must be positive;
	// implementations are free to reject non-positive values via
	// ErrStore(ErrInvalidTTL).
	Add(ctx context.Context, key string, ttl time.Duration) error
}

// ErrStoreFn is the function type for ErrStore.
type ErrStoreFn func(causes ...error) error

// ErrNotFoundFn is the function type for ErrNotFound.
type ErrNotFoundFn func(causes ...error) error
