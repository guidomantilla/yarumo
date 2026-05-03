package sets

import "testing"

func TestStringifyBasic_string(t *testing.T) {
	t.Parallel()

	result := stringifyBasic("hello")

	if result != "hello" {
		t.Fatalf("expected 'hello', got %s", result)
	}
}

func TestStringifyBasic_int(t *testing.T) {
	t.Parallel()

	result := stringifyBasic(42)

	if result != "42" {
		t.Fatalf("expected '42', got %s", result)
	}
}

func TestStringifyBasic_int8(t *testing.T) {
	t.Parallel()

	result := stringifyBasic(int8(8))

	if result != "8" {
		t.Fatalf("expected '8', got %s", result)
	}
}

func TestStringifyBasic_int16(t *testing.T) {
	t.Parallel()

	result := stringifyBasic(int16(16))

	if result != "16" {
		t.Fatalf("expected '16', got %s", result)
	}
}

func TestStringifyBasic_int32(t *testing.T) {
	t.Parallel()

	result := stringifyBasic(int32(32))

	if result != "32" {
		t.Fatalf("expected '32', got %s", result)
	}
}

func TestStringifyBasic_int64(t *testing.T) {
	t.Parallel()

	result := stringifyBasic(int64(64))

	if result != "64" {
		t.Fatalf("expected '64', got %s", result)
	}
}

func TestStringifyBasic_uint(t *testing.T) {
	t.Parallel()

	result := stringifyBasic(uint(10))

	if result != "10" {
		t.Fatalf("expected '10', got %s", result)
	}
}

func TestStringifyBasic_uint8(t *testing.T) {
	t.Parallel()

	result := stringifyBasic(uint8(8))

	if result != "8" {
		t.Fatalf("expected '8', got %s", result)
	}
}

func TestStringifyBasic_uint16(t *testing.T) {
	t.Parallel()

	result := stringifyBasic(uint16(16))

	if result != "16" {
		t.Fatalf("expected '16', got %s", result)
	}
}

func TestStringifyBasic_uint32(t *testing.T) {
	t.Parallel()

	result := stringifyBasic(uint32(32))

	if result != "32" {
		t.Fatalf("expected '32', got %s", result)
	}
}

func TestStringifyBasic_uint64(t *testing.T) {
	t.Parallel()

	result := stringifyBasic(uint64(64))

	if result != "64" {
		t.Fatalf("expected '64', got %s", result)
	}
}

func TestStringifyBasic_float32(t *testing.T) {
	t.Parallel()

	result := stringifyBasic(float32(3.14))

	if result != "3.14" {
		t.Fatalf("expected '3.14', got %s", result)
	}
}

func TestStringifyBasic_float64(t *testing.T) {
	t.Parallel()

	result := stringifyBasic(3.14)

	if result != "3.14" {
		t.Fatalf("expected '3.14', got %s", result)
	}
}

func TestStringifyBasic_bool(t *testing.T) {
	t.Parallel()

	result := stringifyBasic(true)

	if result != "true" {
		t.Fatalf("expected 'true', got %s", result)
	}
}

type customType struct {
	value int
}

func TestStringifyBasic_unknown(t *testing.T) {
	t.Parallel()

	result := stringifyBasic(customType{1})

	if result != "" {
		t.Fatalf("expected empty for unknown type, got %s", result)
	}
}

type stringer struct{}

func (s stringer) String() string {
	return "custom"
}

func TestStringify_stringer(t *testing.T) {
	t.Parallel()

	result := stringify(stringer{})

	if result != "custom" {
		t.Fatalf("expected 'custom', got %s", result)
	}
}

func TestStringify_fallbackToBasic(t *testing.T) {
	t.Parallel()

	result := stringify(42)

	if result != "42" {
		t.Fatalf("expected '42', got %s", result)
	}
}

func TestStringify_unknownFallback(t *testing.T) {
	t.Parallel()

	result := stringify(customType{1})

	if result != "?" {
		t.Fatalf("expected '?', got %s", result)
	}
}
