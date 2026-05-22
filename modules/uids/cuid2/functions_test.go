package cuid2

import "testing"

func TestCUID2(t *testing.T) {
	t.Parallel()

	t.Run("returns non-empty string", func(t *testing.T) {
		t.Parallel()

		got, err := CUID2()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got == "" {
			t.Fatal("expected non-empty string")
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		t.Parallel()

		a, errA := CUID2()
		if errA != nil {
			t.Fatalf("unexpected error: %v", errA)
		}

		b, errB := CUID2()
		if errB != nil {
			t.Fatalf("unexpected error: %v", errB)
		}

		if a == b {
			t.Fatalf("expected unique values, got same: %s", a)
		}
	})
}

func TestIsCUID2(t *testing.T) {
	t.Parallel()

	t.Run("accepts generated CUID2", func(t *testing.T) {
		t.Parallel()

		id, err := CUID2()
		if err != nil {
			t.Fatalf("CUID2: %v", err)
		}

		if !IsCUID2(id) {
			t.Fatalf("IsCUID2(%q) = false, want true", id)
		}
	})

	t.Run("rejects empty string", func(t *testing.T) {
		t.Parallel()

		if IsCUID2("") {
			t.Fatal("IsCUID2(\"\") = true, want false")
		}
	})

	t.Run("rejects wrong length", func(t *testing.T) {
		t.Parallel()

		// 23 chars — one short of the default 24.
		if IsCUID2("abcdefghijklmnopqrstuvw") {
			t.Fatal("IsCUID2(too short) = true, want false")
		}

		// 25 chars — one above the default 24.
		if IsCUID2("abcdefghijklmnopqrstuvwxy") {
			t.Fatal("IsCUID2(too long) = true, want false")
		}
	})

	t.Run("rejects strings starting with a digit", func(t *testing.T) {
		t.Parallel()

		id := "1bcdefghijklmnopqrstuvwx"
		if IsCUID2(id) {
			t.Fatalf("IsCUID2(%q) = true, want false", id)
		}
	})

	t.Run("rejects uppercase characters", func(t *testing.T) {
		t.Parallel()

		id := "Abcdefghijklmnopqrstuvwx"
		if IsCUID2(id) {
			t.Fatalf("IsCUID2(%q) = true, want false", id)
		}
	})
}
