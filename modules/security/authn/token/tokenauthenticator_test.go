package token

import (
	"context"
	"errors"
	"sort"
	"testing"

	ctokens "github.com/guidomantilla/yarumo/crypto/tokens"
	"github.com/guidomantilla/yarumo/security/authn"
)

func newTestMethod(t *testing.T) *ctokens.Method {
	t.Helper()

	method := ctokens.NewMethod("test", ctokens.AlgorithmHS256, ctokens.WithKey([]byte("test-secret-key-32-bytes-long-xxx")))
	return method
}

func mintToken(t *testing.T, method *ctokens.Method, subject string, payload ctokens.Payload) string {
	t.Helper()

	token, err := method.Generate(subject, payload)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	return token
}

func TestNewTokenAuthenticator(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil authenticator", func(t *testing.T) {
		t.Parallel()

		auth := NewTokenAuthenticator(newTestMethod(t))
		if auth == nil {
			t.Fatal("NewTokenAuthenticator returned nil")
		}
	})

	t.Run("with custom claim keys", func(t *testing.T) {
		t.Parallel()

		auth := NewTokenAuthenticator(newTestMethod(t),
			WithSubjectClaim("user_id"),
			WithNameClaim("display"),
			WithRolesClaim("groups"),
		)
		if auth == nil {
			t.Fatal("NewTokenAuthenticator returned nil")
		}
	})
}

func TestTokenAuthenticator_Validate(t *testing.T) {
	t.Parallel()

	t.Run("happy path returns principal", func(t *testing.T) {
		t.Parallel()

		method := newTestMethod(t)
		token := mintToken(t, method, "u-1", ctokens.Payload{
			"sub":   "u-1",
			"name":  "Alice",
			"roles": []any{"admin", "auditor"},
			"tid":   "acme",
		})

		auth := NewTokenAuthenticator(method)

		principal, err := auth.Validate(context.Background(), token)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if principal == nil {
			t.Fatal("Validate returned nil principal")
		}

		if principal.ID != "u-1" {
			t.Fatalf("ID = %q, want %q", principal.ID, "u-1")
		}

		if principal.Name != "Alice" {
			t.Fatalf("Name = %q, want %q", principal.Name, "Alice")
		}

		sort.Strings(principal.Roles)

		if len(principal.Roles) != 2 || principal.Roles[0] != "admin" || principal.Roles[1] != "auditor" {
			t.Fatalf("Roles = %v, want [admin auditor]", principal.Roles)
		}

		tid, ok := principal.Attributes["tid"]
		if !ok || tid != "acme" {
			t.Fatalf("Attributes[tid] = %v ok=%v, want acme/true", tid, ok)
		}
	})

	t.Run("empty token rejected", func(t *testing.T) {
		t.Parallel()

		auth := NewTokenAuthenticator(newTestMethod(t))

		_, err := auth.Validate(context.Background(), "")
		if err == nil {
			t.Fatal("Validate(\"\") returned nil error, want non-nil")
		}

		if !errors.Is(err, authn.ErrTokenEmpty) {
			t.Fatalf("error %v does not match ErrTokenEmpty", err)
		}

		if !errors.Is(err, authn.ErrAuthenticationFailed) {
			t.Fatalf("error %v does not match ErrAuthenticationFailed", err)
		}
	})

	t.Run("invalid token rejected", func(t *testing.T) {
		t.Parallel()

		auth := NewTokenAuthenticator(newTestMethod(t))

		_, err := auth.Validate(context.Background(), "not.a.jwt")
		if err == nil {
			t.Fatal("Validate(garbage) returned nil error, want non-nil")
		}

		if !errors.Is(err, authn.ErrTokenInvalid) {
			t.Fatalf("error %v does not match ErrTokenInvalid", err)
		}
	})

	t.Run("signed with wrong key rejected", func(t *testing.T) {
		t.Parallel()

		good := newTestMethod(t)
		bad := ctokens.NewMethod("bad", ctokens.AlgorithmHS256, ctokens.WithKey([]byte("OTHER-key-32-bytes-long-aaaaaaaa")))

		token := mintToken(t, bad, "u-1", ctokens.Payload{"sub": "u-1"})

		auth := NewTokenAuthenticator(good)

		_, err := auth.Validate(context.Background(), token)
		if err == nil {
			t.Fatal("Validate with wrong-key token returned nil error")
		}

		if !errors.Is(err, authn.ErrTokenInvalid) {
			t.Fatalf("error %v does not match ErrTokenInvalid", err)
		}
	})

	t.Run("missing subject claim rejected", func(t *testing.T) {
		t.Parallel()

		method := newTestMethod(t)
		token := mintToken(t, method, "envelope-sub", ctokens.Payload{
			"name": "Alice",
		})

		auth := NewTokenAuthenticator(method)

		_, err := auth.Validate(context.Background(), token)
		if err == nil {
			t.Fatal("Validate without sub claim returned nil error")
		}

		if !errors.Is(err, ErrSubjectClaimMissing) {
			t.Fatalf("error %v does not match ErrSubjectClaimMissing", err)
		}

		if !errors.Is(err, authn.ErrTokenInvalid) {
			t.Fatalf("error %v does not match ErrTokenInvalid", err)
		}
	})

	t.Run("custom subject claim key", func(t *testing.T) {
		t.Parallel()

		method := newTestMethod(t)
		token := mintToken(t, method, "envelope-sub", ctokens.Payload{
			"user_id": "u-42",
			"name":    "Bob",
		})

		auth := NewTokenAuthenticator(method, WithSubjectClaim("user_id"))

		principal, err := auth.Validate(context.Background(), token)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if principal.ID != "u-42" {
			t.Fatalf("ID = %q, want u-42", principal.ID)
		}
	})

	t.Run("custom roles claim with []string", func(t *testing.T) {
		t.Parallel()

		// crypto/tokens marshals Payload via encoding/json, which
		// turns []string into []any on the way back. To exercise the
		// []string branch we construct the Payload directly without
		// the JSON round-trip via the validateFn override.
		method := ctokens.NewMethod("rolesstr", ctokens.AlgorithmHS256,
			ctokens.WithKey([]byte("k-32-bytes-aaaaaaaaaaaaaaaaaaaaaa")),
			ctokens.WithValidateFn(func(_ *ctokens.Method, _ string) (ctokens.Payload, error) {
				return ctokens.Payload{
					"sub":   "u-7",
					"roles": []string{"viewer"},
				}, nil
			}),
		)

		auth := NewTokenAuthenticator(method)

		principal, err := auth.Validate(context.Background(), "anything")
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if len(principal.Roles) != 1 || principal.Roles[0] != "viewer" {
			t.Fatalf("Roles = %v, want [viewer]", principal.Roles)
		}
	})

	t.Run("non-string roles entries skipped", func(t *testing.T) {
		t.Parallel()

		method := newTestMethod(t)
		token := mintToken(t, method, "u-1", ctokens.Payload{
			"sub":   "u-1",
			"roles": []any{"admin", 42, "viewer"},
		})

		auth := NewTokenAuthenticator(method)

		principal, err := auth.Validate(context.Background(), token)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		want := []string{"admin", "viewer"}
		if len(principal.Roles) != len(want) || principal.Roles[0] != want[0] || principal.Roles[1] != want[1] {
			t.Fatalf("Roles = %v, want %v", principal.Roles, want)
		}
	})

	t.Run("missing roles produces empty non-nil slice", func(t *testing.T) {
		t.Parallel()

		method := newTestMethod(t)
		token := mintToken(t, method, "u-1", ctokens.Payload{
			"sub": "u-1",
		})

		auth := NewTokenAuthenticator(method)

		principal, err := auth.Validate(context.Background(), token)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if principal.Roles == nil {
			t.Fatal("Roles is nil, want empty non-nil slice")
		}

		if len(principal.Roles) != 0 {
			t.Fatalf("Roles = %v, want empty", principal.Roles)
		}
	})

	t.Run("non-string name claim ignored", func(t *testing.T) {
		t.Parallel()

		method := newTestMethod(t)
		token := mintToken(t, method, "u-1", ctokens.Payload{
			"sub":  "u-1",
			"name": 12345,
		})

		auth := NewTokenAuthenticator(method)

		principal, err := auth.Validate(context.Background(), token)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if principal.Name != "" {
			t.Fatalf("Name = %q, want empty (non-string ignored)", principal.Name)
		}
	})
}
