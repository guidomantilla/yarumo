package store

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// StoreType is the error domain identifier for messaging store
// operations.
const StoreType = "messaging-store"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for messaging store failure modes.
var (
	// ErrStoreFailed is the top-level sentinel embedded in every
	// store-domain Error returned by ErrStore.
	ErrStoreFailed = errors.New("store operation failed")
	// ErrStoreNotFound indicates the requested key is not present in
	// the store. Returned by MessageStore.Get when no message has been
	// stored under the given key (or when an in-memory metadata entry
	// has expired and a Get-shaped backend reports the miss the same
	// way).
	ErrStoreNotFound = errors.New("key not found in store")
	// ErrStoreClosed indicates the store is no longer accepting
	// operations because its lifecycle has been stopped. Returned by
	// the in-memory MetadataStore after Stop, so consumers can
	// distinguish "store is dead" from "key just isn't there".
	ErrStoreClosed = errors.New("store is closed")
	// ErrInvalidTTL indicates an Add was attempted with a non-positive
	// TTL. The MetadataStore contract requires a positive TTL —
	// permanent entries should use a payload-bearing store, not a
	// dedup store.
	ErrInvalidTTL = errors.New("ttl must be positive")
)

// Error is the domain error type for messaging store operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("messaging-store %s error: %s", e.Type, e.Err)
}

// ErrStore wraps the given causes into a domain Error joined with
// ErrStoreFailed. Use this factory for generic store failures (Add
// with invalid TTL, Stop-after-use violations, etc.).
func ErrStore(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: StoreType,
			Err:  errors.Join(append(causes, ErrStoreFailed)...),
		},
	}
}

// ErrNotFound wraps the given causes into a domain Error joined with
// ErrStoreNotFound. Returned by MessageStore.Get when the requested
// key is not present.
func ErrNotFound(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: StoreType,
			Err:  errors.Join(append(causes, ErrStoreNotFound)...),
		},
	}
}
