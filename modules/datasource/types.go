// Package datasource provides the workspace-wide contract for relational
// database access: a transport-agnostic Context describing how to reach
// a backend, a Connection interface owning the live backend client, and
// a TransactionHandler that runs a function inside a backend-managed
// transaction.
//
// The package is a pure contract module: it has no external dependencies
// beyond modules/common and performs no I/O. Concrete drivers live in
// sibling modules (for example modules/datasource/gorm) and bring their
// own backend dependencies; each driver implements Connection and
// TransactionHandler. Drivers MAY additionally implement
// lifecycle.Component so they can be wired via lifecycle.Build.
//
// Naming reflects the workspace convention: New<Type> for infallible
// constructors that return interfaces, Build<Type> for fallible
// lifecycle-aware constructors that return (Type, lifecycle.CloseFn,
// error). Drivers are expected to expose a BuildDB-style helper.
//
// The contract intentionally mirrors the
// github.com/guidomantilla/go-feather-lib/pkg/datasource reference
// package adapted to the yarumo conventions (typed errors, getter
// methods, no global state).
package datasource

import (
	"context"
)

// Type compliance vars for the constructors published by this package.
var (
	_ Context        = (*context_)(nil)
	_ NewContextFn   = NewContext
)

// NewContextFn is the function type for NewContext.
type NewContextFn func(url string, user string, password string, server string, service string) Context

// TxFn is the function executed inside a transaction by a
// TransactionHandler. The implementation receives a derived
// context.Context that drivers MAY decorate with the active transaction
// handle (see driver-specific helpers such as gorm.TxFromContext); the
// same ctx must be propagated to every repository call that should
// participate in the transaction.
type TxFn func(ctx context.Context) error

// Context describes how to reach a datasource backend. It is the
// transport-agnostic input consumed by driver constructors that build
// their own DSN string by substituting the placeholders :username,
// :password, :server and :service in Url with the matching getters.
//
// Implementations are intentionally narrow; access is mediated by
// getters so future migrations (sealed types, redaction overrides,
// telemetry-friendly formatting) do not break consumers.
type Context interface {
	// Url returns the DSN template, with placeholders already
	// substituted at construction time.
	Url() string
	// User returns the authentication user.
	User() string
	// Password returns the authentication password.
	Password() string
	// Server returns the server endpoint (host or host:port).
	Server() string
	// Service returns the logical service/database identifier.
	Service() string
}

// Connection is the live backend connection owned by a driver. The
// concrete return type of Connect is driver-specific (for example
// *gorm.DB, *mongo.Client); the cross-driver contract here only
// guarantees that the driver can open, expose and tear down the
// underlying handle.
//
// Connect MUST be idempotent: invoking it more than once returns the
// same handle if already connected, or re-establishes the connection
// when it has been previously closed.
//
// Close releases every resource owned by the connection (sockets, file
// handles, pool goroutines). It MUST be idempotent and safe to call
// concurrently.
type Connection interface {
	// Connect opens (or reuses) the live backend handle and returns it
	// as an opaque any. Drivers expose a typed helper alongside this
	// method (for example connection.DB() *gorm.DB) so callers do not
	// type-assert at every call site.
	Connect(ctx context.Context) (any, error)
	// Close releases the connection's underlying resources. Idempotent.
	Close(ctx context.Context) error
	// Context returns the Context that produced this connection.
	Context() Context
}

// TransactionHandler runs fn inside a backend-managed transaction. The
// driver implementation MUST:
//   - propagate the active transaction handle through ctx so repository
//     calls receive the same transactional handle;
//   - commit when fn returns nil;
//   - rollback when fn returns a non-nil error or panics;
//   - wrap backend errors with the driver's domain error factory so
//     consumers can match via errors.Is against ErrTransactionFailed.
type TransactionHandler interface {
	// HandleTransaction begins a transaction, invokes fn with a derived
	// ctx, and commits or rolls back depending on fn's outcome.
	HandleTransaction(ctx context.Context, fn TxFn) error
}
