package cache

import (
	"testing"
)

func TestJSONCodec_Encode(t *testing.T) {
	t.Parallel()

	t.Run("encodes value to JSON bytes", func(t *testing.T) {
		t.Parallel()

		raw, err := JSONCodec{}.Encode(map[string]int{"a": 1})
		if err != nil {
			t.Fatalf("Encode: %v", err)
		}

		want := `{"a":1}`
		if string(raw) != want {
			t.Fatalf("Encode = %s, want %s", string(raw), want)
		}
	})

	t.Run("returns error when value is not JSON-serializable", func(t *testing.T) {
		t.Parallel()

		_, err := JSONCodec{}.Encode(make(chan int))
		if err == nil {
			t.Fatal("expected error encoding channel")
		}
	})
}

func TestJSONCodec_Decode(t *testing.T) {
	t.Parallel()

	t.Run("decodes JSON bytes into the target", func(t *testing.T) {
		t.Parallel()

		var got map[string]int
		err := JSONCodec{}.Decode([]byte(`{"a":1}`), &got)
		if err != nil {
			t.Fatalf("Decode: %v", err)
		}

		if got["a"] != 1 {
			t.Fatalf("got[a] = %d, want 1", got["a"])
		}
	})

	t.Run("returns error on invalid JSON", func(t *testing.T) {
		t.Parallel()

		var got map[string]int
		err := JSONCodec{}.Decode([]byte("not-json"), &got)
		if err == nil {
			t.Fatal("expected error decoding invalid JSON")
		}
	})
}
