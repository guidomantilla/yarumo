package probability

import (
	"strings"
	"testing"
)

func TestDistribution_String_empty(t *testing.T) {
	t.Parallel()

	d := Distribution{}

	if d.String() != "{}" {
		t.Fatalf("expected {}, got %q", d.String())
	}
}

func TestDistribution_String_single(t *testing.T) {
	t.Parallel()

	d := Distribution{"heads": 1.0}
	result := d.String()

	if result != "{heads=1}" {
		t.Fatalf("expected {heads=1}, got %q", result)
	}
}

func TestDistribution_String_multiple(t *testing.T) {
	t.Parallel()

	d := Distribution{"b": 0.3, "a": 0.7}
	result := d.String()

	// Should be sorted by key.
	if !strings.Contains(result, "a=") || !strings.Contains(result, "b=") {
		t.Fatalf("expected both outcomes, got %q", result)
	}

	if !strings.HasPrefix(result, "{a=") {
		t.Fatalf("expected sorted output starting with a, got %q", result)
	}
}

func TestCPT_String_noParents(t *testing.T) {
	t.Parallel()

	c := NewCPT("Rain", nil)
	c.Set(Assignment{}, Distribution{"true": 0.2, "false": 0.8})

	result := c.String()

	if !strings.HasPrefix(result, "CPT(Rain)") {
		t.Fatalf("expected CPT(Rain), got %q", result)
	}
}

func TestCPT_String_withParents(t *testing.T) {
	t.Parallel()

	c := NewCPT("WetGrass", []Var{"Rain", "Sprinkler"})

	result := c.String()

	if !strings.Contains(result, "CPT(WetGrass | Rain, Sprinkler)") {
		t.Fatalf("expected parent list, got %q", result)
	}
}

func TestCPT_String_withEntries(t *testing.T) {
	t.Parallel()

	c := NewCPT("X", []Var{"Y"})
	c.Set(Assignment{"Y": "y1"}, Distribution{"a": 0.5, "b": 0.5})

	result := c.String()

	if !strings.Contains(result, "Y=y1") {
		t.Fatalf("expected entry key, got %q", result)
	}
}

func TestCPT_String_emptyEntries(t *testing.T) {
	t.Parallel()

	c := NewCPT("X", nil)
	result := c.String()

	if result != "CPT(X)" {
		t.Fatalf("expected CPT(X), got %q", result)
	}
}

func TestFactor_String_empty(t *testing.T) {
	t.Parallel()

	f := NewFactor(nil, nil)
	result := f.String()

	if result != "Factor()" {
		t.Fatalf("expected Factor(), got %q", result)
	}
}

func TestFactor_String_withVars(t *testing.T) {
	t.Parallel()

	f := NewFactor([]Var{"A", "B"}, map[string]Prob{
		"A=t,B=t": 0.3,
	})

	result := f.String()

	if !strings.HasPrefix(result, "Factor(A, B)") {
		t.Fatalf("expected Factor(A, B), got %q", result)
	}

	if !strings.Contains(result, "A=t,B=t") {
		t.Fatalf("expected table entry, got %q", result)
	}
}

func TestFactor_String_emptyKeyEntry(t *testing.T) {
	t.Parallel()

	f := NewFactor(nil, map[string]Prob{
		"": 0.5,
	})

	result := f.String()

	if !strings.Contains(result, "0.5") {
		t.Fatalf("expected value, got %q", result)
	}
}

func TestCPT_String_emptyKeyEntry(t *testing.T) {
	t.Parallel()

	c := NewCPT("Rain", nil)
	c.Set(Assignment{}, Distribution{"true": 0.5, "false": 0.5})

	result := c.String()

	if !strings.Contains(result, "true=0.5") {
		t.Fatalf("expected distribution, got %q", result)
	}
}
