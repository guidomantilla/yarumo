package tokens

import (
    "errors"
    "testing"
)

func TestFakeGenerator_Generate_SuccessAndValidations(t *testing.T) {
    out := "ok"
    g := &FakeGenerator{
        GenerateFn: func(subject string, principal Principal) (*string, error) {
            if subject != "sub" || principal["a"] != 1 {
                t.Fatalf("unexpected args: %v %v", subject, principal)
            }
            return &out, nil
        },
    }
    tok, err := g.Generate("sub", Principal{"a": 1})
    if err != nil || tok == nil || *tok != "ok" {
        t.Fatalf("unexpected: %v %v", tok, err)
    }

    // subject vacío
    _, err = g.Generate("", Principal{"a": 1})
    if err == nil || !errors.Is(err, ErrTokenGenerationFailed) {
        t.Fatalf("expected wrapped generation error for empty subject, got %v", err)
    }

    // principal vacío
    _, err = g.Generate("sub", nil)
    if err == nil || !errors.Is(err, ErrTokenGenerationFailed) {
        t.Fatalf("expected wrapped generation error for empty principal, got %v", err)
    }
}

func TestFakeGenerator_Validate_SuccessAndValidations(t *testing.T) {
    g := &FakeGenerator{
        ValidateFn: func(tokenString string) (Principal, error) {
            if tokenString != "abc" {
                t.Fatalf("unexpected token: %s", tokenString)
            }
            return Principal{"x": 2}, nil
        },
    }
    p, err := g.Validate("abc")
    if err != nil || p["x"].(int) != 2 {
        t.Fatalf("unexpected: %v %v", p, err)
    }

    // token vacío
    _, err = g.Validate("")
    if err == nil || !errors.Is(err, ErrTokenValidationFailed) {
        t.Fatalf("expected wrapped validation error for empty token, got %v", err)
    }
}
