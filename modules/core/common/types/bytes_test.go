package types

import "testing"

func TestBytes_ToHex(t *testing.T) {
	t.Parallel()

	t.Run("nil bytes", func(t *testing.T) {
		t.Parallel()

		got := Bytes(nil).ToHex()
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("empty bytes", func(t *testing.T) {
		t.Parallel()

		got := Bytes{}.ToHex()
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("ascii text", func(t *testing.T) {
		t.Parallel()

		got := Bytes("hello").ToHex()
		if got != "68656c6c6f" {
			t.Fatalf("got %q, want %q", got, "68656c6c6f")
		}
	})

	t.Run("binary data", func(t *testing.T) {
		t.Parallel()

		got := Bytes([]byte{251, 255, 239}).ToHex()
		if got != "fbffef" {
			t.Fatalf("got %q, want %q", got, "fbffef")
		}
	})
}

func TestBytes_ToBase64Std(t *testing.T) {
	t.Parallel()

	t.Run("nil bytes", func(t *testing.T) {
		t.Parallel()

		got := Bytes(nil).ToBase64Std()
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("empty bytes", func(t *testing.T) {
		t.Parallel()

		got := Bytes{}.ToBase64Std()
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("ascii text with padding", func(t *testing.T) {
		t.Parallel()

		got := Bytes("hello").ToBase64Std()
		if got != "aGVsbG8=" {
			t.Fatalf("got %q, want %q", got, "aGVsbG8=")
		}
	})

	t.Run("binary with plus and slash", func(t *testing.T) {
		t.Parallel()

		got := Bytes([]byte{251, 255, 239}).ToBase64Std()
		if got != "+//v" {
			t.Fatalf("got %q, want %q", got, "+//v")
		}
	})
}

func TestBytes_ToBase64RawStd(t *testing.T) {
	t.Parallel()

	t.Run("nil bytes", func(t *testing.T) {
		t.Parallel()

		got := Bytes(nil).ToBase64RawStd()
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("empty bytes", func(t *testing.T) {
		t.Parallel()

		got := Bytes{}.ToBase64RawStd()
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("ascii text without padding", func(t *testing.T) {
		t.Parallel()

		got := Bytes("hello").ToBase64RawStd()
		if got != "aGVsbG8" {
			t.Fatalf("got %q, want %q", got, "aGVsbG8")
		}
	})

	t.Run("binary with plus and slash", func(t *testing.T) {
		t.Parallel()

		got := Bytes([]byte{251, 255, 239}).ToBase64RawStd()
		if got != "+//v" {
			t.Fatalf("got %q, want %q", got, "+//v")
		}
	})
}

func TestBytes_ToBase64Url(t *testing.T) {
	t.Parallel()

	t.Run("nil bytes", func(t *testing.T) {
		t.Parallel()

		got := Bytes(nil).ToBase64Url()
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("empty bytes", func(t *testing.T) {
		t.Parallel()

		got := Bytes{}.ToBase64Url()
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("ascii text with padding", func(t *testing.T) {
		t.Parallel()

		got := Bytes("hello").ToBase64Url()
		if got != "aGVsbG8=" {
			t.Fatalf("got %q, want %q", got, "aGVsbG8=")
		}
	})

	t.Run("binary uses url-safe chars", func(t *testing.T) {
		t.Parallel()

		got := Bytes([]byte{251, 255, 239}).ToBase64Url()
		if got != "-__v" {
			t.Fatalf("got %q, want %q", got, "-__v")
		}
	})
}

func TestBytes_ToBase64RawUrl(t *testing.T) {
	t.Parallel()

	t.Run("nil bytes", func(t *testing.T) {
		t.Parallel()

		got := Bytes(nil).ToBase64RawUrl()
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("empty bytes", func(t *testing.T) {
		t.Parallel()

		got := Bytes{}.ToBase64RawUrl()
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("ascii text without padding", func(t *testing.T) {
		t.Parallel()

		got := Bytes("hello").ToBase64RawUrl()
		if got != "aGVsbG8" {
			t.Fatalf("got %q, want %q", got, "aGVsbG8")
		}
	})

	t.Run("binary uses url-safe chars", func(t *testing.T) {
		t.Parallel()

		got := Bytes([]byte{251, 255, 239}).ToBase64RawUrl()
		if got != "-__v" {
			t.Fatalf("got %q, want %q", got, "-__v")
		}
	})
}
