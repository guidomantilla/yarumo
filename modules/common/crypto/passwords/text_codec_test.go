package passwords

import (
	"encoding/json"
	"errors"
	"testing"
)

const textCodecTestName = "Argon2"

func TestMethod_MarshalText(t *testing.T) {
	t.Parallel()

	t.Run("returns the registry name", func(t *testing.T) {
		t.Parallel()

		data, err := Argon2.MarshalText()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(data) != textCodecTestName {
			t.Fatalf("expected %q, got %q", textCodecTestName, string(data))
		}
	})

	t.Run("panics on nil receiver", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic on nil receiver, got none")
			}
		}()

		var m *Method
		_, _ = m.MarshalText()
	})
}

func TestMethod_UnmarshalText(t *testing.T) {
	t.Parallel()

	t.Run("resolves predefined name and overwrites receiver", func(t *testing.T) {
		t.Parallel()

		m := new(Method)

		err := m.UnmarshalText([]byte(textCodecTestName))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if m.Name() != textCodecTestName {
			t.Fatalf("expected %q, got %q", textCodecTestName, m.Name())
		}
	})

	t.Run("returns ErrAlgorithmNotSupported for unknown name", func(t *testing.T) {
		t.Parallel()

		m := new(Method)

		err := m.UnmarshalText([]byte("BOGUS"))
		if err == nil {
			t.Fatal("expected error for unknown algorithm, got nil")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error via errors.As, got %T", err)
		}

		if domErr.Type != PasswordMethod {
			t.Fatalf("expected type %q, got %q", PasswordMethod, domErr.Type)
		}
	})
}

func TestMethod_TextCodec_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("MarshalText then UnmarshalText preserves identity", func(t *testing.T) {
		t.Parallel()

		data, err := Argon2.MarshalText()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := new(Method)

		err = got.UnmarshalText(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != Argon2.Name() {
			t.Fatalf("expected %q, got %q", Argon2.Name(), got.Name())
		}

		if got.prefix != Argon2.prefix {
			t.Fatalf("expected prefix %q, got %q", Argon2.prefix, got.prefix)
		}
	})
}

func TestMethod_JSON_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("encoding/json honors TextMarshaler/TextUnmarshaler", func(t *testing.T) {
		t.Parallel()

		type Config struct {
			Encoder *Method `json:"encoder"`
		}

		in := Config{Encoder: Argon2}

		raw, err := json.Marshal(in)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		const expected = `{"encoder":"Argon2"}`
		if string(raw) != expected {
			t.Fatalf("expected %q, got %q", expected, string(raw))
		}

		var out Config

		err = json.Unmarshal(raw, &out)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if out.Encoder == nil {
			t.Fatal("expected non-nil Encoder after unmarshal")
		}

		if out.Encoder.Name() != textCodecTestName {
			t.Fatalf("expected %q, got %q", textCodecTestName, out.Encoder.Name())
		}
	})
}
