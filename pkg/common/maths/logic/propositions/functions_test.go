package propositions

import "testing"

func TestNNF_DeMorgan(t *testing.T) {
	A := V("A")
	B := V("B")
	f := NotF{F: AndF{L: A, R: B}}
	got := ToNNF(f)
	want := OrF{L: A.Not(), R: B.Not()}
	if !Equivalent(got, want) {
		t.Fatalf("NNF DeMorgan fallo: got=%s want=%s", got.String(), want.String())
	}
}

func TestCNF_Distribute(t *testing.T) {
	A := V("A")
	B := V("B")
	C := V("C")
	f := OrF{L: A, R: AndF{L: B, R: C}}
	got := ToCNF(f)
	want := AndF{
		L: OrF{L: A, R: B},
		R: OrF{L: A, R: C},
	}
	if !Equivalent(got, want) {
		t.Fatalf("CNF distribucion fallo: got=%s want=%s", got.String(), want.String())
	}
}

func TestDNF_Distribute(t *testing.T) {
	A := V("A")
	B := V("B")
	C := V("C")
	f := AndF{L: A, R: OrF{L: B, R: C}}
	got := ToDNF(f)
	want := OrF{
		L: AndF{L: A, R: B},
		R: AndF{L: A, R: C},
	}
	if !Equivalent(got, want) {
		t.Fatalf("DNF distribucion fallo: got=%s want=%s", got.String(), want.String())
	}
}
