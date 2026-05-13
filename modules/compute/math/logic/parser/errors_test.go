package parser

import (
	"errors"
	"testing"
)

func TestErrParse(t *testing.T) {
	t.Parallel()

	t.Run("returns *ParseError", func(t *testing.T) {
		t.Parallel()

		err := ErrParse(0, 0, "test", ErrEmptyInput)

		var pe *ParseError
		if !errors.As(err, &pe) {
			t.Fatalf("expected *ParseError, got %T", err)
		}
	})

	t.Run("Error returns the message", func(t *testing.T) {
		t.Parallel()

		pe := ErrParse(5, 6, "boom", ErrUnexpectedToken)
		if pe.Error() != "boom" {
			t.Fatalf("expected %q, got %q", "boom", pe.Error())
		}
	})

	t.Run("fields are set correctly", func(t *testing.T) {
		t.Parallel()

		pe := ErrParse(3, 7, "msg", ErrUnclosedParen)
		if pe.Pos != 3 {
			t.Fatalf("expected Pos 3, got %d", pe.Pos)
		}
		if pe.Col != 7 {
			t.Fatalf("expected Col 7, got %d", pe.Col)
		}
		if pe.Msg != "msg" {
			t.Fatalf("expected Msg 'msg', got %s", pe.Msg)
		}
	})

	t.Run("Type is parser type", func(t *testing.T) {
		t.Parallel()

		pe := ErrParse(0, 0, "test", ErrEmptyInput)
		if pe.Type != ParserType {
			t.Fatalf("expected type %s, got %s", ParserType, pe.Type)
		}
	})

	t.Run("Unwrap returns cause", func(t *testing.T) {
		t.Parallel()

		pe := ErrParse(0, 0, "test", ErrEmptyInput)
		if !errors.Is(pe, ErrEmptyInput) {
			t.Fatal("expected ErrEmptyInput in chain")
		}
	})

	t.Run("zero args still wraps ErrParseFailed", func(t *testing.T) {
		t.Parallel()

		pe := ErrParse(0, 0, "test")
		if !errors.Is(pe, ErrParseFailed) {
			t.Fatal("expected ErrParseFailed in chain")
		}
	})

	t.Run("multiple causes joined", func(t *testing.T) {
		t.Parallel()

		pe := ErrParse(0, 0, "test", ErrEmptyInput, ErrUnexpectedToken)
		if !errors.Is(pe, ErrEmptyInput) {
			t.Fatal("expected ErrEmptyInput in chain")
		}
		if !errors.Is(pe, ErrUnexpectedToken) {
			t.Fatal("expected ErrUnexpectedToken in chain")
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrUnexpectedEnd", func(t *testing.T) {
		t.Parallel()
		if ErrUnexpectedEnd.Error() != "unexpected end of input" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrUnexpectedToken", func(t *testing.T) {
		t.Parallel()
		if ErrUnexpectedToken.Error() != "unexpected token" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrUnclosedParen", func(t *testing.T) {
		t.Parallel()
		if ErrUnclosedParen.Error() != "unclosed parenthesis" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrEmptyInput", func(t *testing.T) {
		t.Parallel()
		if ErrEmptyInput.Error() != "empty input" {
			t.Fatal("unexpected message")
		}
	})

	t.Run("ErrParseFailed", func(t *testing.T) {
		t.Parallel()
		if ErrParseFailed.Error() != "parser operation failed" {
			t.Fatal("unexpected message")
		}
	})
}
