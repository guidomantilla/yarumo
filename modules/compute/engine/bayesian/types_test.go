package bayesian

import (
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/stats"
)

func TestCPT(t *testing.T) {
	t.Parallel()

	c := CPT{
		Variable: "WetGrass",
		Parents:  []stats.Var{"Rain"},
		Entries:  make(map[string]stats.Distribution),
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
		Variables: []stats.Var{"A", "B"},
		Table:     map[string]stats.Prob{"A=t,B=t": 0.3},
	}

	if len(f.Variables) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(f.Variables))
	}

	if f.Table["A=t,B=t"] != 0.3 {
		t.Fatalf("expected 0.3, got %f", float64(f.Table["A=t,B=t"]))
	}
}
