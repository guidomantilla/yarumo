package datasource

import (
	"strings"
	"testing"
)

func TestNewContext(t *testing.T) {
	t.Parallel()

	t.Run("substitutes the placeholders in the url", func(t *testing.T) {
		t.Parallel()

		ctx := NewContext(
			"postgres://:username::password@:server/:service",
			"alice",
			"s3cret",
			"localhost:5432",
			"app",
		)

		got := ctx.Url()
		want := "postgres://alice:s3cret@localhost:5432/app"

		if got != want {
			t.Fatalf("Url = %q, want %q", got, want)
		}
	})

	t.Run("returns the structured fields verbatim", func(t *testing.T) {
		t.Parallel()

		ctx := NewContext("url", "u", "p", "s", "svc")

		if ctx.User() != "u" {
			t.Fatalf("User = %q, want %q", ctx.User(), "u")
		}

		if ctx.Password() != "p" {
			t.Fatalf("Password = %q, want %q", ctx.Password(), "p")
		}

		if ctx.Server() != "s" {
			t.Fatalf("Server = %q, want %q", ctx.Server(), "s")
		}

		if ctx.Service() != "svc" {
			t.Fatalf("Service = %q, want %q", ctx.Service(), "svc")
		}
	})

	t.Run("leaves the url unchanged when no placeholders are present", func(t *testing.T) {
		t.Parallel()

		ctx := NewContext("file::memory:", "u", "p", "s", "svc")

		if strings.Contains(ctx.Url(), ":username") {
			t.Fatalf("did not expect placeholders to remain")
		}

		if ctx.Url() != "file::memory:" {
			t.Fatalf("Url = %q, want %q", ctx.Url(), "file::memory:")
		}
	})
}
