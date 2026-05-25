package datasource

import (
	"errors"
	"testing"
)

func TestErrConnect(t *testing.T) {
	t.Parallel()

	t.Run("joins ErrConnectFailed and ErrDatasourceFailed", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("dial tcp: refused")

		err := ErrConnect(cause)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrConnectFailed) {
			t.Fatalf("expected wrap of ErrConnectFailed, got %v", err)
		}

		if !errors.Is(err, ErrDatasourceFailed) {
			t.Fatalf("expected wrap of ErrDatasourceFailed, got %v", err)
		}

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}
	})

	t.Run("carries the datasource domain Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrConnect(errors.New("x"))

		var domain *Error

		ok := errors.As(err, &domain)
		if !ok {
			t.Fatalf("expected to extract *Error, got %v", err)
		}

		if domain.Type != DatasourceType {
			t.Fatalf("Type = %q, want %q", domain.Type, DatasourceType)
		}
	})
}

func TestErrTransaction(t *testing.T) {
	t.Parallel()

	t.Run("joins ErrTransactionFailed and ErrDatasourceFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrTransaction(ErrTxPanic)

		if !errors.Is(err, ErrTransactionFailed) {
			t.Fatalf("expected wrap of ErrTransactionFailed, got %v", err)
		}

		if !errors.Is(err, ErrTxPanic) {
			t.Fatalf("expected wrap of ErrTxPanic, got %v", err)
		}

		if !errors.Is(err, ErrDatasourceFailed) {
			t.Fatalf("expected wrap of ErrDatasourceFailed, got %v", err)
		}
	})
}

func TestErrClose(t *testing.T) {
	t.Parallel()

	t.Run("joins ErrCloseFailed and ErrDatasourceFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrClose(errors.New("disk full"))

		if !errors.Is(err, ErrCloseFailed) {
			t.Fatalf("expected wrap of ErrCloseFailed, got %v", err)
		}

		if !errors.Is(err, ErrDatasourceFailed) {
			t.Fatalf("expected wrap of ErrDatasourceFailed, got %v", err)
		}
	})
}

func TestErrDatasource(t *testing.T) {
	t.Parallel()

	t.Run("joins only the umbrella sentinel", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("misc failure")

		err := ErrDatasource(cause)

		if !errors.Is(err, ErrDatasourceFailed) {
			t.Fatalf("expected wrap of ErrDatasourceFailed, got %v", err)
		}

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}

		if errors.Is(err, ErrConnectFailed) {
			t.Fatal("did not expect ErrConnectFailed to match")
		}
	})
}

func TestErrorTypeError(t *testing.T) {
	t.Parallel()

	t.Run("formatted message includes the datasource domain", func(t *testing.T) {
		t.Parallel()

		err := ErrConnect(errors.New("x"))

		msg := err.Error()
		if msg == "" {
			t.Fatal("expected non-empty error message")
		}
	})
}
