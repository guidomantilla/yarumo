package probability

import (
	"testing"
)

func TestVar(t *testing.T) {
	t.Parallel()

	v := Var("Rain")
	if string(v) != "Rain" {
		t.Fatalf("expected Rain, got %s", string(v))
	}
}

func TestOutcome(t *testing.T) {
	t.Parallel()

	o := Outcome("true")
	if string(o) != "true" {
		t.Fatalf("expected true, got %s", string(o))
	}
}

func TestProb(t *testing.T) {
	t.Parallel()

	p := Prob(0.75)
	if float64(p) != 0.75 {
		t.Fatalf("expected 0.75, got %f", float64(p))
	}
}

func TestDistribution(t *testing.T) {
	t.Parallel()

	d := Distribution{
		"heads": 0.5,
		"tails": 0.5,
	}

	if d["heads"] != 0.5 {
		t.Fatalf("expected 0.5 for heads, got %f", float64(d["heads"]))
	}
}

func TestAssignment(t *testing.T) {
	t.Parallel()

	a := Assignment{
		"Rain":     "true",
		"Sprinkle": "false",
	}

	if a["Rain"] != "true" {
		t.Fatalf("expected true, got %s", string(a["Rain"]))
	}
}

func TestCPT(t *testing.T) {
	t.Parallel()

	c := CPT{
		Variable: "WetGrass",
		Parents:  []Var{"Rain"},
		Entries:  make(map[string]Distribution),
	}

	if c.Variable != "WetGrass" {
		t.Fatalf("expected WetGrass, got %s", string(c.Variable))
	}

	if len(c.Parents) != 1 {
		t.Fatalf("expected 1 parent, got %d", len(c.Parents))
	}
}

func TestFactor(t *testing.T) {
	t.Parallel()

	f := Factor{
		Variables: []Var{"A", "B"},
		Table:     map[string]Prob{"A=t,B=t": 0.3},
	}

	if len(f.Variables) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(f.Variables))
	}

	if f.Table["A=t,B=t"] != 0.3 {
		t.Fatalf("expected 0.3, got %f", float64(f.Table["A=t,B=t"]))
	}
}
