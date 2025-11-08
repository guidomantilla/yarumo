package entailment

import (
	"testing"

	p "github.com/guidomantilla/yarumo/deprecated/logic/propositions"
)

func TestEntails_ModusPonens(t *testing.T) {
	A, B := p.V("A"), p.V("B")
	kb := []p.Formula{
		p.ImplF{L: A, R: B},
		A,
	}
	if !Entails(kb, B) {
		t.Fatalf("KB should entail B by modus ponens")
	}
}

func TestEntailsWithCounterModel_NotEntailed(t *testing.T) {
	A, B := p.V("A"), p.V("B")
	kb := []p.Formula{
		p.ImplF{L: A, R: B},
	}
	res, err := EntailsWithCounterModel(kb, B)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Entails {
		t.Fatalf("expected not entailed")
	}
	if len(res.CounterModel) == 0 {
		t.Fatalf("expected countermodel assignment")
	}
}
