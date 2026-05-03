package expressions

import (
	"testing"
)

func TestOpKind_Symbol(t *testing.T) {
	t.Parallel()

	t.Run("returns + for OpAdd", func(t *testing.T) {
		t.Parallel()
		if got := OpAdd.Symbol(); got != "+" {
			t.Fatalf("expected +, got %s", got)
		}
	})

	t.Run("returns - for OpSub", func(t *testing.T) {
		t.Parallel()
		if got := OpSub.Symbol(); got != "-" {
			t.Fatalf("expected -, got %s", got)
		}
	})

	t.Run("returns * for OpMul", func(t *testing.T) {
		t.Parallel()
		if got := OpMul.Symbol(); got != "*" {
			t.Fatalf("expected *, got %s", got)
		}
	})

	t.Run("returns / for OpDiv", func(t *testing.T) {
		t.Parallel()
		if got := OpDiv.Symbol(); got != "/" {
			t.Fatalf("expected /, got %s", got)
		}
	})

	t.Run("returns %% for OpMod", func(t *testing.T) {
		t.Parallel()
		if got := OpMod.Symbol(); got != "%" {
			t.Fatalf("expected %%, got %s", got)
		}
	})

	t.Run("returns == for OpEq", func(t *testing.T) {
		t.Parallel()
		if got := OpEq.Symbol(); got != "==" {
			t.Fatalf("expected ==, got %s", got)
		}
	})

	t.Run("returns != for OpNeq", func(t *testing.T) {
		t.Parallel()
		if got := OpNeq.Symbol(); got != "!=" {
			t.Fatalf("expected !=, got %s", got)
		}
	})

	t.Run("returns < for OpLt", func(t *testing.T) {
		t.Parallel()
		if got := OpLt.Symbol(); got != "<" {
			t.Fatalf("expected <, got %s", got)
		}
	})

	t.Run("returns <= for OpLte", func(t *testing.T) {
		t.Parallel()
		if got := OpLte.Symbol(); got != "<=" {
			t.Fatalf("expected <=, got %s", got)
		}
	})

	t.Run("returns > for OpGt", func(t *testing.T) {
		t.Parallel()
		if got := OpGt.Symbol(); got != ">" {
			t.Fatalf("expected >, got %s", got)
		}
	})

	t.Run("returns >= for OpGte", func(t *testing.T) {
		t.Parallel()
		if got := OpGte.Symbol(); got != ">=" {
			t.Fatalf("expected >=, got %s", got)
		}
	})

	t.Run("returns - for OpNeg", func(t *testing.T) {
		t.Parallel()
		if got := OpNeg.Symbol(); got != "-" {
			t.Fatalf("expected -, got %s", got)
		}
	})

	t.Run("returns ! for OpNot", func(t *testing.T) {
		t.Parallel()
		if got := OpNot.Symbol(); got != "!" {
			t.Fatalf("expected !, got %s", got)
		}
	})

	t.Run("returns ? for unknown OpKind", func(t *testing.T) {
		t.Parallel()
		if got := OpKind(99).Symbol(); got != "?" {
			t.Fatalf("expected ?, got %s", got)
		}
	})
}
