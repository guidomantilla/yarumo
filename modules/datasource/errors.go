package datasource

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// DatasourceType is the error domain identifier for datasource errors.
// Driver sub-modules SHOULD define their own *Type constant (for example
// "datasource-gorm") and wrap the sentinels exported here so callers can
// match either the driver domain or the cross-driver sentinel via
// errors.Is.
const DatasourceType = "datasource"

var (
	_ error            = (*Error)(nil)
	_ ErrConnectFn     = ErrConnect
	_ ErrTransactionFn = ErrTransaction
	_ ErrCloseFn       = ErrClose
	_ ErrDatasourceFn  = ErrDatasource
)

// Error is the domain error type for cross-driver datasource failures.
// Drivers wrap their backend-specific errors via Err<Operation>
// factories that embed their own cerrs.TypedError.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for cross-driver datasource failure modes. Drivers
// join the sentinel that matches the operation that failed; consumers
// can use errors.Is to branch on the operation regardless of which
// backend produced the error.
var (
	// ErrDatasourceFailed is the umbrella sentinel joined by every
	// datasource domain error.
	ErrDatasourceFailed = errors.New("datasource operation failed")
	// ErrConnectFailed is joined when establishing a connection to the
	// backend fails (DNS, dial, TLS handshake, authentication).
	ErrConnectFailed = errors.New("datasource connect failed")
	// ErrTransactionFailed is joined when a Begin/Commit/Rollback step
	// of a transaction fails, or when the user-supplied TxFn returns an
	// error.
	ErrTransactionFailed = errors.New("datasource transaction failed")
	// ErrCloseFailed is joined when releasing the connection resources
	// fails.
	ErrCloseFailed = errors.New("datasource close failed")
	// ErrConnectionNil is joined when a nil Connection is passed to an
	// API that requires a live connection.
	ErrConnectionNil = errors.New("connection is nil")
	// ErrTxFnNil is joined when a TransactionHandler is invoked with a
	// nil TxFn.
	ErrTxFnNil = errors.New("transaction function is nil")
	// ErrTxPanic is joined when the user-supplied TxFn panics. The
	// panic value is recovered by the driver, formatted, and added to
	// the error chain so the panic is surfaced through standard error
	// matching.
	ErrTxPanic = errors.New("transaction function panicked")
)

// ErrConnectFn is the function type for ErrConnect.
type ErrConnectFn func(causes ...error) error

// ErrTransactionFn is the function type for ErrTransaction.
type ErrTransactionFn func(causes ...error) error

// ErrCloseFn is the function type for ErrClose.
type ErrCloseFn func(causes ...error) error

// ErrDatasourceFn is the function type for ErrDatasource.
type ErrDatasourceFn func(causes ...error) error

// ErrConnect creates a datasource domain error joining the given causes
// with ErrConnectFailed and the umbrella ErrDatasourceFailed.
func ErrConnect(causes ...error) error {
	return newError(append(causes, ErrConnectFailed)...)
}

// ErrTransaction creates a datasource domain error joining the given
// causes with ErrTransactionFailed and the umbrella ErrDatasourceFailed.
func ErrTransaction(causes ...error) error {
	return newError(append(causes, ErrTransactionFailed)...)
}

// ErrClose creates a datasource domain error joining the given causes
// with ErrCloseFailed and the umbrella ErrDatasourceFailed.
func ErrClose(causes ...error) error {
	return newError(append(causes, ErrCloseFailed)...)
}

// ErrDatasource creates a generic datasource domain error joining the
// given causes with the umbrella ErrDatasourceFailed sentinel. Intended
// as the catch-all factory for failures that do not fit any of the
// specific Err<Operation> factories.
func ErrDatasource(causes ...error) error {
	return newError(causes...)
}

// newError is the shared internal constructor used by every public
// factory. It always joins ErrDatasourceFailed at the tail so consumers
// can match the umbrella sentinel via errors.Is regardless of the
// specific factory used.
func newError(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: DatasourceType,
			Err:  errors.Join(append(causes, ErrDatasourceFailed)...),
		},
	}
}
