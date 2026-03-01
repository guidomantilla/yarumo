package variable

import "testing"

func TestNewOptions_defaults(t *testing.T) {
	t.Parallel()

	o := NewOptions()

	if o.resolution != defaultResolution {
		t.Fatalf("expected default resolution %d, got %d", defaultResolution, o.resolution)
	}
}

func TestWithResolution(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithResolution(200))

	if o.resolution != 200 {
		t.Fatalf("expected 200, got %d", o.resolution)
	}
}

func TestWithResolution_zero(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithResolution(0))

	if o.resolution != defaultResolution {
		t.Fatalf("expected default, got %d", o.resolution)
	}
}

func TestWithResolution_negative(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithResolution(-5))

	if o.resolution != defaultResolution {
		t.Fatalf("expected default, got %d", o.resolution)
	}
}
