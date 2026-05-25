package gorm

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	cdatasource "github.com/guidomantilla/yarumo/datasource"
)

func TestNewTransactionHandler(t *testing.T) {
	t.Parallel()

	t.Run("returns a non-nil handler", func(t *testing.T) {
		t.Parallel()

		conn := newTestConn(t)

		h := NewTransactionHandler(conn)
		if h == nil {
			t.Fatal("expected non-nil handler")
		}
	})
}

func TestHandleTransaction_Commit(t *testing.T) {
	t.Parallel()

	t.Run("commits when fn returns nil", func(t *testing.T) {
		t.Parallel()

		conn := newTestConn(t)
		gdb := mustMigrate(t, conn)

		h := NewTransactionHandler(conn)

		err := h.HandleTransaction(context.Background(), func(ctx context.Context) error {
			tx, ok := TxFromContext(ctx)
			if !ok {
				t.Fatal("expected tx in ctx")
			}

			return tx.Create(&item{Name: "alpha"}).Error
		})

		if err != nil {
			t.Fatalf("HandleTransaction: %v", err)
		}

		var got int64
		_ = gdb.Model(&item{}).Count(&got).Error

		if got != 1 {
			t.Fatalf("Count after commit = %d, want 1", got)
		}
	})
}

func TestHandleTransaction_Rollback(t *testing.T) {
	t.Parallel()

	t.Run("rolls back on fn error", func(t *testing.T) {
		t.Parallel()

		conn := newTestConn(t)
		gdb := mustMigrate(t, conn)

		h := NewTransactionHandler(conn)
		boom := errors.New("boom")

		err := h.HandleTransaction(context.Background(), func(ctx context.Context) error {
			tx, _ := TxFromContext(ctx)

			createErr := tx.Create(&item{Name: "rollback-me"}).Error
			if createErr != nil {
				return createErr
			}

			return boom
		})

		if !errors.Is(err, boom) {
			t.Fatalf("expected wrap of boom, got %v", err)
		}

		if !errors.Is(err, cdatasource.ErrTransactionFailed) {
			t.Fatalf("expected wrap of datasource.ErrTransactionFailed, got %v", err)
		}

		var got int64
		_ = gdb.Model(&item{}).Count(&got).Error

		if got != 0 {
			t.Fatalf("Count after rollback = %d, want 0", got)
		}
	})
}

func TestHandleTransaction_Panic(t *testing.T) {
	t.Parallel()

	t.Run("rolls back on panic and surfaces ErrTxPanic", func(t *testing.T) {
		t.Parallel()

		conn := newTestConn(t)
		gdb := mustMigrate(t, conn)

		h := NewTransactionHandler(conn)

		err := h.HandleTransaction(context.Background(), func(ctx context.Context) error {
			tx, _ := TxFromContext(ctx)

			_ = tx.Create(&item{Name: "panic-me"}).Error

			panic("kaboom")
		})

		if !errors.Is(err, cdatasource.ErrTxPanic) {
			t.Fatalf("expected wrap of datasource.ErrTxPanic, got %v", err)
		}

		if !errors.Is(err, cdatasource.ErrTransactionFailed) {
			t.Fatalf("expected wrap of datasource.ErrTransactionFailed, got %v", err)
		}

		var got int64
		_ = gdb.Model(&item{}).Count(&got).Error

		if got != 0 {
			t.Fatalf("Count after panic = %d, want 0", got)
		}
	})
}

func TestHandleTransaction_NilFn(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrTxFnNil when fn is nil", func(t *testing.T) {
		t.Parallel()

		conn := newTestConn(t)

		h := NewTransactionHandler(conn)

		err := h.HandleTransaction(context.Background(), nil)
		if !errors.Is(err, cdatasource.ErrTxFnNil) {
			t.Fatalf("expected wrap of ErrTxFnNil, got %v", err)
		}
	})
}

func TestTxFromContext(t *testing.T) {
	t.Parallel()

	t.Run("returns false for nil context", func(t *testing.T) {
		t.Parallel()

		var nilCtx context.Context

		tx, ok := TxFromContext(nilCtx)
		if ok || tx != nil {
			t.Fatalf("expected (nil, false), got (%v, %v)", tx, ok)
		}
	})

	t.Run("returns false when no tx is installed", func(t *testing.T) {
		t.Parallel()

		tx, ok := TxFromContext(context.Background())
		if ok || tx != nil {
			t.Fatalf("expected (nil, false), got (%v, %v)", tx, ok)
		}
	})

	t.Run("returns the installed tx", func(t *testing.T) {
		t.Parallel()

		conn := newTestConn(t)

		raw, err := conn.Connect(context.Background())
		if err != nil {
			t.Fatalf("Connect: %v", err)
		}

		gdb, ok := raw.(*gorm.DB)
		if !ok {
			t.Fatal("expected *gorm.DB")
		}

		ctx := contextWithTx(context.Background(), gdb)

		tx, found := TxFromContext(ctx)
		if !found {
			t.Fatal("expected to find tx")
		}

		if tx != gdb {
			t.Fatal("expected installed tx to be returned unchanged")
		}
	})
}
