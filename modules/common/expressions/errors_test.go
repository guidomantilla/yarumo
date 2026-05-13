package expressions

import (
	"errors"
	"strings"
	"testing"
)

func TestParseError(t *testing.T) {
	t.Parallel()

	t.Run("implements error interface", func(t *testing.T) {
		t.Parallel()
		var err error = ErrParse(5, 10, "bad token", ErrUnexpectedToken)
		if err == nil {
			t.Fatal("expected non-nil error")
		}
	})

	t.Run("Error includes position and message", func(t *testing.T) {
		t.Parallel()
		pe := ErrParse(5, 10, "bad token", ErrUnexpectedToken)
		got := pe.Error()
		if !strings.Contains(got, "5") {
			t.Fatalf("expected position in error, got %s", got)
		}
		if !strings.Contains(got, "bad token") {
			t.Fatalf("expected message in error, got %s", got)
		}
	})

	t.Run("Unwrap returns cause", func(t *testing.T) {
		t.Parallel()
		pe := ErrParse(0, 1, "test", ErrEmptyInput)
		if !errors.Is(pe, ErrEmptyInput) {
			t.Fatal("expected ErrEmptyInput in chain")
		}
	})

	t.Run("fields are set correctly", func(t *testing.T) {
		t.Parallel()
		pe := ErrParse(3, 7, "msg", ErrUnclosedString)
		if pe.Pos != 3 {
			t.Fatalf("expected Pos 3, got %d", pe.Pos)
		}
		if pe.End != 7 {
			t.Fatalf("expected End 7, got %d", pe.End)
		}
		if pe.Msg != "msg" {
			t.Fatalf("expected Msg 'msg', got %s", pe.Msg)
		}
	})

	t.Run("Type is expression", func(t *testing.T) {
		t.Parallel()
		pe := ErrParse(0, 0, "test")
		if pe.Type != ExpressionType {
			t.Fatalf("expected type %s, got %s", ExpressionType, pe.Type)
		}
	})

	t.Run("multiple causes joined", func(t *testing.T) {
		t.Parallel()
		pe := ErrParse(0, 1, "test", ErrEmptyInput, ErrUnexpectedToken)
		if !errors.Is(pe, ErrEmptyInput) {
			t.Fatal("expected ErrEmptyInput in chain")
		}
		if !errors.Is(pe, ErrUnexpectedToken) {
			t.Fatal("expected ErrUnexpectedToken in chain")
		}
	})

	t.Run("zero args still wraps ErrParseFailed", func(t *testing.T) {
		t.Parallel()
		pe := ErrParse(0, 0, "test")
		if !errors.Is(pe, ErrParseFailed) {
			t.Fatal("expected ErrParseFailed in chain")
		}
	})
}

func TestEvalError(t *testing.T) {
	t.Parallel()

	t.Run("implements error interface", func(t *testing.T) {
		t.Parallel()
		var err error = ErrEval("bad eval", ErrTypeMismatch)
		if err == nil {
			t.Fatal("expected non-nil error")
		}
	})

	t.Run("Error includes message", func(t *testing.T) {
		t.Parallel()
		ee := ErrEval("division by zero", ErrDivisionByZero)
		got := ee.Error()
		if !strings.Contains(got, "division by zero") {
			t.Fatalf("expected message in error, got %s", got)
		}
	})

	t.Run("Unwrap returns cause", func(t *testing.T) {
		t.Parallel()
		ee := ErrEval("test", ErrDivisionByZero)
		if !errors.Is(ee, ErrDivisionByZero) {
			t.Fatal("expected ErrDivisionByZero in chain")
		}
	})

	t.Run("Type is expression", func(t *testing.T) {
		t.Parallel()
		ee := ErrEval("test")
		if ee.Type != ExpressionType {
			t.Fatalf("expected type %s, got %s", ExpressionType, ee.Type)
		}
	})

	t.Run("Msg field is set", func(t *testing.T) {
		t.Parallel()
		ee := ErrEval("custom msg", ErrNilAccess)
		if ee.Msg != "custom msg" {
			t.Fatalf("expected 'custom msg', got %s", ee.Msg)
		}
	})

	t.Run("zero args still wraps ErrEvalFailed", func(t *testing.T) {
		t.Parallel()
		ee := ErrEval("test")
		if !errors.Is(ee, ErrEvalFailed) {
			t.Fatal("expected ErrEvalFailed in chain")
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrEmptyInput", func(t *testing.T) {
		t.Parallel()
		if ErrEmptyInput.Error() != "empty input" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrUnexpectedToken", func(t *testing.T) {
		t.Parallel()
		if ErrUnexpectedToken.Error() != "unexpected token" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrUnexpectedEnd", func(t *testing.T) {
		t.Parallel()
		if ErrUnexpectedEnd.Error() != "unexpected end of input" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrUnclosedParen", func(t *testing.T) {
		t.Parallel()
		if ErrUnclosedParen.Error() != "unclosed parenthesis" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrUnclosedBracket", func(t *testing.T) {
		t.Parallel()
		if ErrUnclosedBracket.Error() != "unclosed bracket" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrUnclosedString", func(t *testing.T) {
		t.Parallel()
		if ErrUnclosedString.Error() != "unclosed string" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrInvalidNumber", func(t *testing.T) {
		t.Parallel()
		if ErrInvalidNumber.Error() != "invalid number" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrTypeMismatch", func(t *testing.T) {
		t.Parallel()
		if ErrTypeMismatch.Error() != "type mismatch" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrDivisionByZero", func(t *testing.T) {
		t.Parallel()
		if ErrDivisionByZero.Error() != "division by zero" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrUnknownField", func(t *testing.T) {
		t.Parallel()
		if ErrUnknownField.Error() != "unknown field" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrUnknownFunc", func(t *testing.T) {
		t.Parallel()
		if ErrUnknownFunc.Error() != "unknown function" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrArgCount", func(t *testing.T) {
		t.Parallel()
		if ErrArgCount.Error() != "wrong argument count" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrNilAccess", func(t *testing.T) {
		t.Parallel()
		if ErrNilAccess.Error() != "nil access" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrParseFailed", func(t *testing.T) {
		t.Parallel()
		if ErrParseFailed.Error() != "expression parse failed" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrEvalFailed", func(t *testing.T) {
		t.Parallel()
		if ErrEvalFailed.Error() != "expression eval failed" {
			t.Fatal("unexpected message")
		}
	})
}
