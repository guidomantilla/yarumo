package tokens

import (
	"errors"
	"strings"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

func TestTokenError_Error(t *testing.T) {
	base := errors.New("root cause")
	terr := &TokenError{TypedError: cerrs.TypedError{Type: TokenGenerationType, Err: base}}
	msg := terr.Error()
	if !strings.Contains(msg, "token generation error:") || !strings.Contains(msg, "root cause") {
		t.Fatalf("unexpected error message: %s", msg)
	}
}

func TestErrTokenGenerationAndValidation(t *testing.T) {
	e1 := errors.New("e1")
	e2 := errors.New("e2")
	g := ErrTokenGeneration(e1, e2)
	if g == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(g.Error(), "token generation error:") || !strings.Contains(g.Error(), ErrTokenGenerationFailed.Error()) {
		t.Fatalf("unexpected generation error: %v", g)
	}

	v := ErrTokenValidation(e1, e2)
	if v == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(v.Error(), "token validation error:") || !strings.Contains(v.Error(), ErrTokenValidationFailed.Error()) {
		t.Fatalf("unexpected validation error: %v", v)
	}
}
