package gorm

import (
	"errors"
	"testing"

	cdatasource "github.com/guidomantilla/yarumo/datasource"
)

func TestErrOpen(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrOpenFailed and cross-driver connect sentinels", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("dial refused")

		err := ErrOpen(cause)

		if !errors.Is(err, ErrOpenFailed) {
			t.Fatalf("expected wrap of ErrOpenFailed, got %v", err)
		}

		if !errors.Is(err, cdatasource.ErrConnectFailed) {
			t.Fatalf("expected wrap of datasource.ErrConnectFailed, got %v", err)
		}

		if !errors.Is(err, cdatasource.ErrDatasourceFailed) {
			t.Fatalf("expected wrap of datasource.ErrDatasourceFailed, got %v", err)
		}

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}
	})
}

func TestErrClose(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrCloseFailed and ErrDatasourceFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrClose(errors.New("x"))

		if !errors.Is(err, cdatasource.ErrCloseFailed) {
			t.Fatalf("expected wrap of datasource.ErrCloseFailed, got %v", err)
		}

		if !errors.Is(err, cdatasource.ErrDatasourceFailed) {
			t.Fatalf("expected wrap of datasource.ErrDatasourceFailed, got %v", err)
		}
	})
}

func TestErrTransaction(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrTransactionFailed and ErrDatasourceFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrTransaction(cdatasource.ErrTxPanic)

		if !errors.Is(err, cdatasource.ErrTransactionFailed) {
			t.Fatalf("expected wrap of datasource.ErrTransactionFailed, got %v", err)
		}

		if !errors.Is(err, cdatasource.ErrTxPanic) {
			t.Fatalf("expected wrap of datasource.ErrTxPanic, got %v", err)
		}
	})
}

func TestErrSQLDB(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrSQLDBUnavailable and ErrDatasourceFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrSQLDB(errors.New("x"))

		if !errors.Is(err, ErrSQLDBUnavailable) {
			t.Fatalf("expected wrap of ErrSQLDBUnavailable, got %v", err)
		}

		if !errors.Is(err, cdatasource.ErrDatasourceFailed) {
			t.Fatalf("expected wrap of datasource.ErrDatasourceFailed, got %v", err)
		}
	})

	t.Run("works with no causes", func(t *testing.T) {
		t.Parallel()

		err := ErrSQLDB()

		if !errors.Is(err, ErrSQLDBUnavailable) {
			t.Fatalf("expected wrap of ErrSQLDBUnavailable, got %v", err)
		}
	})
}
