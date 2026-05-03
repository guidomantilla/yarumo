package rules

import "testing"

func TestNewOptions_defaults(t *testing.T) {
	t.Parallel()

	o := NewOptions()

	if o.operator != And {
		t.Fatalf("expected And, got %d", o.operator)
	}

	if o.weight != 1.0 {
		t.Fatalf("expected 1.0, got %f", o.weight)
	}
}

func TestWithOperator_and(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithOperator(And))

	if o.operator != And {
		t.Fatalf("expected And, got %d", o.operator)
	}
}

func TestWithOperator_or(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithOperator(Or))

	if o.operator != Or {
		t.Fatalf("expected Or, got %d", o.operator)
	}
}

func TestWithOperator_outOfRange(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithOperator(Operator(99)))

	if o.operator != And {
		t.Fatalf("expected default And, got %d", o.operator)
	}
}

func TestWithWeight(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithWeight(0.5))

	if o.weight != 0.5 {
		t.Fatalf("expected 0.5, got %f", o.weight)
	}
}

func TestWithWeight_zero(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithWeight(0.0))

	if o.weight != 0.0 {
		t.Fatalf("expected 0.0, got %f", o.weight)
	}
}

func TestWithWeight_outOfRange_negative(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithWeight(-0.1))

	if o.weight != 1.0 {
		t.Fatalf("expected default 1.0, got %f", o.weight)
	}
}

func TestWithWeight_outOfRange_above(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithWeight(1.1))

	if o.weight != 1.0 {
		t.Fatalf("expected default 1.0, got %f", o.weight)
	}
}
