package fuzzy

import "testing"

func TestDegree(t *testing.T) {
	t.Parallel()

	d := Degree(0.75)
	if float64(d) != 0.75 {
		t.Fatalf("expected 0.75, got %f", float64(d))
	}
}

func TestSet(t *testing.T) {
	t.Parallel()

	s := Set{Name: "cold", Fn: Constant(0.5)}

	if s.Name != "cold" {
		t.Fatalf("expected cold, got %s", s.Name)
	}

	if s.Fn(10) != 0.5 {
		t.Fatalf("expected 0.5, got %f", float64(s.Fn(10)))
	}
}

func TestPoint(t *testing.T) {
	t.Parallel()

	p := Point{X: 3.5, Degree: 0.8}

	if p.X != 3.5 {
		t.Fatalf("expected 3.5, got %f", p.X)
	}

	if p.Degree != 0.8 {
		t.Fatalf("expected 0.8, got %f", float64(p.Degree))
	}
}

func TestMembershipFn(t *testing.T) {
	t.Parallel()

	fn := MembershipFn(func(x float64) Degree { return Degree(x / 10) })
	result := fn(5)

	if result != 0.5 {
		t.Fatalf("expected 0.5, got %f", float64(result))
	}
}

func TestTNormFn(t *testing.T) {
	t.Parallel()

	fn := TNormFn(Min)
	result := fn(0.3, 0.7)

	if result != 0.3 {
		t.Fatalf("expected 0.3, got %f", float64(result))
	}
}

func TestTConormFn(t *testing.T) {
	t.Parallel()

	fn := TConormFn(Max)
	result := fn(0.3, 0.7)

	if result != 0.7 {
		t.Fatalf("expected 0.7, got %f", float64(result))
	}
}

func TestDefuzzifyFn(t *testing.T) {
	t.Parallel()

	fn := DefuzzifyFn(Centroid)
	result := fn([]float64{1, 2, 3}, []Degree{0, 1, 0})

	if result != 2.0 {
		t.Fatalf("expected 2.0, got %f", result)
	}
}
