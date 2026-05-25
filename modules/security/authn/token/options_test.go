package token_test

import (
	"testing"

	authntoken "github.com/guidomantilla/yarumo/security/authn/token"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()

		opts := authntoken.NewOptions()
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})

	t.Run("with all overrides", func(t *testing.T) {
		t.Parallel()

		opts := authntoken.NewOptions(
			authntoken.WithSubjectClaim("uid"),
			authntoken.WithNameClaim("display"),
			authntoken.WithRolesClaim("scopes"),
		)
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}

func TestWithSubjectClaim(t *testing.T) {
	t.Parallel()

	t.Run("empty is ignored", func(t *testing.T) {
		t.Parallel()

		opts := authntoken.NewOptions(authntoken.WithSubjectClaim(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}

func TestWithNameClaim(t *testing.T) {
	t.Parallel()

	t.Run("empty is ignored", func(t *testing.T) {
		t.Parallel()

		opts := authntoken.NewOptions(authntoken.WithNameClaim(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}

func TestWithRolesClaim(t *testing.T) {
	t.Parallel()

	t.Run("empty is ignored", func(t *testing.T) {
		t.Parallel()

		opts := authntoken.NewOptions(authntoken.WithRolesClaim(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}
