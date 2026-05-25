package jwt_test

import (
	"testing"

	authnjwt "github.com/guidomantilla/yarumo/security/authn/jwt"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()

		opts := authnjwt.NewOptions()
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})

	t.Run("with all overrides", func(t *testing.T) {
		t.Parallel()

		opts := authnjwt.NewOptions(
			authnjwt.WithSubjectClaim("uid"),
			authnjwt.WithNameClaim("display"),
			authnjwt.WithRolesClaim("scopes"),
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

		opts := authnjwt.NewOptions(authnjwt.WithSubjectClaim(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}

func TestWithNameClaim(t *testing.T) {
	t.Parallel()

	t.Run("empty is ignored", func(t *testing.T) {
		t.Parallel()

		opts := authnjwt.NewOptions(authnjwt.WithNameClaim(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}

func TestWithRolesClaim(t *testing.T) {
	t.Parallel()

	t.Run("empty is ignored", func(t *testing.T) {
		t.Parallel()

		opts := authnjwt.NewOptions(authnjwt.WithRolesClaim(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}
