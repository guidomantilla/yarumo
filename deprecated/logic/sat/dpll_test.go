package sat

import (
	"testing"

	p "github.com/guidomantilla/yarumo/deprecated/logic/propositions"
)

func TestDPLL_Satisfiable(t *testing.T) {
	A, B := p.V("A"), p.V("B")
	// (A ∨ B) ∧ (¬A ∨ B) ∧ (A ∨ ¬B)
	f := p.AndF{L: p.OrF{L: A, R: B}, R: p.AndF{L: p.OrF{L: A.Not(), R: B}, R: p.OrF{L: A, R: B.Not()}}}
	cnf, err := FromFormulaToCNF(f)
	if err != nil {
		t.Fatalf("CNF conversion error: %v", err)
	}
	ok, asg := DPLL(cnf, nil)
	if !ok {
		t.Fatalf("expected satisfiable, got UNSAT")
	}
	if len(asg) == 0 {
		t.Fatalf("expected non-empty assignment")
	}
}

func TestDPLL_Unsatisfiable(t *testing.T) {
	A := p.V("A")
	// A ∧ ¬A
	f := p.AndF{L: A, R: p.NotF{F: A}}
	cnf, err := FromFormulaToCNF(f)
	if err != nil {
		t.Fatalf("CNF conversion error: %v", err)
	}
	ok, _ := DPLL(cnf, nil)
	if ok {
		t.Fatalf("expected UNSAT, got SAT")
	}
}
