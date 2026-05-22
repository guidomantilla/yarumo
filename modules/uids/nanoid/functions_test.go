package nanoid

import "testing"

func TestNANOID(t *testing.T) {
	t.Parallel()

	t.Run("returns non-empty string", func(t *testing.T) {
		t.Parallel()

		got, err := NANOID()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got == "" {
			t.Fatal("expected non-empty string")
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		t.Parallel()

		a, errA := NANOID()
		if errA != nil {
			t.Fatalf("unexpected error: %v", errA)
		}

		b, errB := NANOID()
		if errB != nil {
			t.Fatalf("unexpected error: %v", errB)
		}

		if a == b {
			t.Fatalf("expected unique values, got same: %s", a)
		}
	})
}

func TestIsNanoID(t *testing.T) {
	t.Parallel()

	t.Run("accepts generated NanoID", func(t *testing.T) {
		t.Parallel()

		id, err := NANOID()
		if err != nil {
			t.Fatalf("NANOID: %v", err)
		}

		if !IsNanoID(id) {
			t.Fatalf("IsNanoID(%q) = false, want true", id)
		}
	})

	t.Run("rejects empty string", func(t *testing.T) {
		t.Parallel()

		if IsNanoID("") {
			t.Fatal("IsNanoID(\"\") = true, want false")
		}
	})

	t.Run("rejects wrong length", func(t *testing.T) {
		t.Parallel()

		// 20 chars — one short of the default 21.
		if IsNanoID("abcdefghij0123456789") {
			t.Fatal("IsNanoID(too short) = true, want false")
		}

		// 22 chars — one above the default 21.
		if IsNanoID("abcdefghij012345678901") {
			t.Fatal("IsNanoID(too long) = true, want false")
		}
	})

	t.Run("rejects characters outside URL-safe alphabet", func(t *testing.T) {
		t.Parallel()

		// 21 chars containing '!' which is not in the URL-safe alphabet.
		id := "abcdefghij0123456789!"
		if IsNanoID(id) {
			t.Fatalf("IsNanoID(%q) = true, want false", id)
		}
	})
}
