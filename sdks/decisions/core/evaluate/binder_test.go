package evaluate

import (
	"testing"

	cexpressions "github.com/guidomantilla/yarumo/core/common/expressions"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/math/logic"
)

type testFullBinder struct{}

func (b testFullBinder) BindDeductive(_ string) logic.Fact {
	return logic.Fact{"a": true}
}

func (b testFullBinder) BindBayesian(_ string) evidence.EvidenceBase {
	return evidence.NewEvidenceBase()
}

func (b testFullBinder) BindFuzzy(_ string) map[string]float64 {
	return map[string]float64{"x": 0.5}
}

func (b testFullBinder) BindExpression(_ string) cexpressions.Context {
	return cexpressions.Context{"x": 1}
}

// Verify Binder interface compliance.
var _ Binder[string] = testFullBinder{}

var _ DeductiveBinder[string] = testFullBinder{}

var _ BayesianBinder[string] = testFullBinder{}

var _ FuzzyBinder[string] = testFullBinder{}

var _ ExpressionBinder[string] = testFullBinder{}

func TestBinder_Interfaces(t *testing.T) {
	t.Parallel()

	b := testFullBinder{}

	t.Run("deductive", func(t *testing.T) {
		t.Parallel()

		facts := b.BindDeductive("test")
		if len(facts) != 1 {
			t.Fatalf("expected 1 fact, got %d", len(facts))
		}
	})

	t.Run("bayesian", func(t *testing.T) {
		t.Parallel()

		eb := b.BindBayesian("test")
		if eb == nil {
			t.Fatal("expected non-nil evidence base")
		}
	})

	t.Run("fuzzy", func(t *testing.T) {
		t.Parallel()

		inputs := b.BindFuzzy("test")
		if len(inputs) != 1 {
			t.Fatalf("expected 1 input, got %d", len(inputs))
		}
	})

	t.Run("expression", func(t *testing.T) {
		t.Parallel()

		ctx := b.BindExpression("test")
		if len(ctx) != 1 {
			t.Fatalf("expected 1 context entry, got %d", len(ctx))
		}
	})
}
