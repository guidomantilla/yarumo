package gorm

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cdatasource "github.com/guidomantilla/yarumo/datasource"
)

// transactionHandler is the GORM-backed datasource.TransactionHandler.
// It opens a transaction via gorm.DB.Transaction, exposes the active
// *gorm.DB to fn through ctx, and recovers panics so callers see a
// rolled-back transaction surfaced as ErrTransaction wrapping
// datasource.ErrTxPanic.
type transactionHandler struct {
	connection cdatasource.Connection
}

// NewTransactionHandler builds a TransactionHandler bound to the given
// Connection. The Connection MUST be the GORM driver's connection (any
// other implementation will trigger a typed nil-handle error on first
// invocation).
func NewTransactionHandler(conn cdatasource.Connection) cdatasource.TransactionHandler {
	cassert.NotNil(conn, "connection is nil")

	return &transactionHandler{connection: conn}
}

// HandleTransaction opens a transaction against the bound Connection
// and invokes fn with a derived ctx carrying the active *gorm.DB.
//
// Semantics:
//   - A nil fn returns ErrTransaction wrapping datasource.ErrTxFnNil.
//   - Failures opening the connection surface as ErrTransaction
//     wrapping the underlying connect error.
//   - A non-nil error from fn triggers Rollback and is returned wrapped
//     via ErrTransaction.
//   - A panic in fn is recovered, the transaction is rolled back, and
//     the panic value is surfaced via ErrTransaction joined with
//     datasource.ErrTxPanic.
func (h *transactionHandler) HandleTransaction(ctx context.Context, fn cdatasource.TxFn) error {
	cassert.NotNil(h, "transaction handler is nil")

	if fn == nil {
		return ErrTransaction(cdatasource.ErrTxFnNil)
	}

	gdb, err := h.openGorm(ctx)
	if err != nil {
		return ErrTransaction(err)
	}

	txErr := gdb.WithContext(ctx).Transaction(func(tx *gorm.DB) (txFnErr error) {
		defer func() {
			r := recover()
			if r != nil {
				txFnErr = ErrTransaction(fmt.Errorf("%w: %v", cdatasource.ErrTxPanic, r))
			}
		}()

		txCtx := contextWithTx(ctx, tx)

		fnErr := fn(txCtx)
		if fnErr != nil {
			return ErrTransaction(fnErr)
		}

		return nil
	})

	if txErr == nil {
		return nil
	}

	// Avoid double-wrapping when txErr is already one of our domain
	// Error values (panic path, or fn returned a wrapped error from
	// our recovery branch).
	var domain *Error

	ok := errors.As(txErr, &domain)
	if ok {
		return txErr
	}

	return ErrTransaction(txErr)
}

// openGorm asks the bound Connection for the live *gorm.DB. The
// indirection lets the handler accept any Connection while still
// requiring the gorm driver in practice.
func (h *transactionHandler) openGorm(ctx context.Context) (*gorm.DB, error) {
	raw, err := h.connection.Connect(ctx)
	if err != nil {
		return nil, err
	}

	gdb, ok := raw.(*gorm.DB)
	if !ok || gdb == nil {
		return nil, ErrSQLDB()
	}

	return gdb, nil
}

// contextWithTx returns a child context carrying tx.
func contextWithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

// TxFromContext extracts the active transactional *gorm.DB previously
// installed by HandleTransaction. The boolean is false when ctx
// carries no active transaction, in which case callers should fall
// back to the connection's DB(ctx) handle.
func TxFromContext(ctx context.Context) (*gorm.DB, bool) {
	if ctx == nil {
		return nil, false
	}

	v := ctx.Value(txCtxKey{})

	tx, ok := v.(*gorm.DB)
	if !ok || tx == nil {
		return nil, false
	}

	return tx, true
}
