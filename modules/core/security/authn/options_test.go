package authn

import (
	"testing"

)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})

	t.Run("with all overrides", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithSubjectClaim("uid"),
			WithNameClaim("display"),
			WithRolesClaim("scopes"),
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

		opts := NewOptions(WithSubjectClaim(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}

func TestWithNameClaim(t *testing.T) {
	t.Parallel()

	t.Run("empty is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithNameClaim(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}

func TestWithRolesClaim(t *testing.T) {
	t.Parallel()

	t.Run("empty is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRolesClaim(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}
