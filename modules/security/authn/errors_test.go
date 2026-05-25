package authn_test

import (
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/security/authn"
)

func TestErrAuthentication(t *testing.T) {
	t.Parallel()

	t.Run("joins ErrAuthenticationFailed", func(t *testing.T) {
		t.Parallel()

		err := authn.ErrAuthentication(authn.ErrTokenInvalid)

		if !errors.Is(err, authn.ErrAuthenticationFailed) {
			t.Fatal("errors.Is(err, ErrAuthenticationFailed) = false, want true")
		}

		if !errors.Is(err, authn.ErrTokenInvalid) {
			t.Fatal("errors.Is(err, ErrTokenInvalid) = false, want true")
		}
	})

	t.Run("includes type in message", func(t *testing.T) {
		t.Parallel()

		err := authn.ErrAuthentication(authn.ErrTokenEmpty)

		msg := err.Error()
		if msg == "" {
			t.Fatal("Error() returned empty string")
		}

		// The "authn" type prefix should always be present.
		want := "authn"

		found := false

		for i := 0; i+len(want) <= len(msg); i++ {
			if msg[i:i+len(want)] == want {
				found = true

				break
			}
		}

		if !found {
			t.Fatalf("Error() = %q, want to contain %q", msg, want)
		}
	})

	t.Run("no causes still joins sentinel", func(t *testing.T) {
		t.Parallel()

		err := authn.ErrAuthentication()
		if !errors.Is(err, authn.ErrAuthenticationFailed) {
			t.Fatal("errors.Is(err, ErrAuthenticationFailed) = false, want true")
		}
	})
}
