package examples

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic/entailment"
	"github.com/guidomantilla/yarumo/maths/logic/parser"
	p "github.com/guidomantilla/yarumo/maths/logic/props"
)

func TestEntails_ModusPonens(t *testing.T) {
	kb := []p.Formula{
		parser.MustParse("A => B"),
		parser.MustParse("A"),
	}
	phi := parser.MustParse("B")
	if !entailment.Entails(kb, phi) {
		t.Fatalf("KB should entail phi (modus ponens)")
	}
}

func TestEntailsWithCounterModel_EntailedAndNotEntailed(t *testing.T) {
	// Entailed case: {A=>B, A} ⊨ B
	kb1 := []p.Formula{parser.MustParse("A => B"), parser.MustParse("A")}
	phi1 := parser.MustParse("B")
	ent1, model1 := entailment.EntailsWithCounterModel(kb1, phi1)
	if !ent1 {
		t.Fatalf("expected entailment to hold, got false with model=%v", model1)
	}
	if model1 != nil && len(model1) != 0 {
		t.Fatalf("expected no countermodel when entailment holds, got %v", model1)
	}

	// Not entailed case: {A=>B} ⊭ C. Expect a countermodel for (A=>B) ∧ ¬C
	kb2 := []p.Formula{parser.MustParse("A => B")}
	phi2 := parser.MustParse("C")
	ent2, model2 := entailment.EntailsWithCounterModel(kb2, phi2)
	if ent2 {
		t.Fatalf("expected entailment NOT to hold")
	}
	if len(model2) == 0 {
		t.Fatalf("expected a non-empty countermodel assignment")
	}
	// Validate the countermodel actually satisfies KB ∧ ¬phi
	target := p.AndF{L: andAllTest(kb2), R: p.NotF{F: phi2}}
	facts := p.Fact(model2)
	if !target.Eval(facts) {
		t.Fatalf("returned assignment does not satisfy KB ∧ ¬phi: %v", model2)
	}
}

// andAllTest is a tiny helper mirroring entailment.andAll for test scope.
func andAllTest(fs []p.Formula) p.Formula {
	if len(fs) == 0 {
		return p.TrueF{}
	}
	acc := fs[0]
	for i := 1; i < len(fs); i++ {
		acc = p.AndF{L: acc, R: fs[i]}
	}
	return acc
}
