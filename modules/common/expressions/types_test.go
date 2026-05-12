package expressions

import (
	"testing"
)

func TestOpKind_Symbol(t *testing.T) {
	t.Parallel()

	t.Run("returns + for OpAdd", func(t *testing.T) {
		t.Parallel()
		got := OpAdd.Symbol()
		if got != "+" {
			t.Fatalf("expected +, got %s", got)
		}
	})

	t.Run("returns - for OpSub", func(t *testing.T) {
		t.Parallel()
		got := OpSub.Symbol()
		if got != "-" {
			t.Fatalf("expected -, got %s", got)
		}
	})

	t.Run("returns * for OpMul", func(t *testing.T) {
		t.Parallel()
		got := OpMul.Symbol()
		if got != "*" {
			t.Fatalf("expected *, got %s", got)
		}
	})

	t.Run("returns / for OpDiv", func(t *testing.T) {
		t.Parallel()
		got := OpDiv.Symbol()
		if got != "/" {
			t.Fatalf("expected /, got %s", got)
		}
	})

	t.Run("returns %% for OpMod", func(t *testing.T) {
		t.Parallel()
		got := OpMod.Symbol()
		if got != "%" {
			t.Fatalf("expected %%, got %s", got)
		}
	})

	t.Run("returns == for OpEq", func(t *testing.T) {
		t.Parallel()
		got := OpEq.Symbol()
		if got != "==" {
			t.Fatalf("expected ==, got %s", got)
		}
	})

	t.Run("returns != for OpNeq", func(t *testing.T) {
		t.Parallel()
		got := OpNeq.Symbol()
		if got != "!=" {
			t.Fatalf("expected !=, got %s", got)
		}
	})

	t.Run("returns < for OpLt", func(t *testing.T) {
		t.Parallel()
		got := OpLt.Symbol()
		if got != "<" {
			t.Fatalf("expected <, got %s", got)
		}
	})

	t.Run("returns <= for OpLte", func(t *testing.T) {
		t.Parallel()
		got := OpLte.Symbol()
		if got != "<=" {
			t.Fatalf("expected <=, got %s", got)
		}
	})

	t.Run("returns > for OpGt", func(t *testing.T) {
		t.Parallel()
		got := OpGt.Symbol()
		if got != ">" {
			t.Fatalf("expected >, got %s", got)
		}
	})

	t.Run("returns >= for OpGte", func(t *testing.T) {
		t.Parallel()
		got := OpGte.Symbol()
		if got != ">=" {
			t.Fatalf("expected >=, got %s", got)
		}
	})

	t.Run("returns - for OpNeg", func(t *testing.T) {
		t.Parallel()
		got := OpNeg.Symbol()
		if got != "-" {
			t.Fatalf("expected -, got %s", got)
		}
	})

	t.Run("returns ! for OpNot", func(t *testing.T) {
		t.Parallel()
		got := OpNot.Symbol()
		if got != "!" {
			t.Fatalf("expected !, got %s", got)
		}
	})

	t.Run("returns ? for unknown OpKind", func(t *testing.T) {
		t.Parallel()
		got := OpKind(99).Symbol()
		if got != "?" {
			t.Fatalf("expected ?, got %s", got)
		}
	})
}
