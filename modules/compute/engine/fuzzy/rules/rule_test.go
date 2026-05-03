package rules

import "testing"

func TestNewRule(t *testing.T) {
	t.Parallel()

	r := NewRule("r1",
		[]Condition{{Variable: "temp", Term: "hot"}},
		Consequent{Variable: "speed", Term: "fast"},
	)

	if r.Name() != "r1" {
		t.Fatalf("expected r1, got %s", r.Name())
	}

	if len(r.Conditions()) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(r.Conditions()))
	}

	if r.Operator() != And {
		t.Fatalf("expected And, got %d", r.Operator())
	}

	if r.Consequent().Variable != "speed" {
		t.Fatalf("expected speed, got %s", r.Consequent().Variable)
	}

	if r.Consequent().Term != "fast" {
		t.Fatalf("expected fast, got %s", r.Consequent().Term)
	}

	if r.Weight() != 1.0 {
		t.Fatalf("expected 1.0, got %f", r.Weight())
	}
}

func TestNewRule_withOptions(t *testing.T) {
	t.Parallel()

	r := NewRule("r2",
		[]Condition{
			{Variable: "temp", Term: "hot"},
			{Variable: "humidity", Term: "high"},
		},
		Consequent{Variable: "speed", Term: "fast"},
		WithOperator(Or),
		WithWeight(0.8),
	)

	if r.Operator() != Or {
		t.Fatalf("expected Or, got %d", r.Operator())
	}

	if r.Weight() != 0.8 {
		t.Fatalf("expected 0.8, got %f", r.Weight())
	}

	if len(r.Conditions()) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(r.Conditions()))
	}
}

func TestNewRule_conditionsDefensiveCopy(t *testing.T) {
	t.Parallel()

	conditions := []Condition{{Variable: "temp", Term: "hot"}}
	r := NewRule("r1", conditions, Consequent{Variable: "speed", Term: "fast"})

	changed := "changed"
	conditions[0].Variable = changed

	if r.Conditions()[0].Variable == changed {
		t.Fatal("expected defensive copy of conditions")
	}
}

func TestRule_Conditions_defensiveCopy(t *testing.T) {
	t.Parallel()

	r := NewRule("r1",
		[]Condition{{Variable: "temp", Term: "hot"}},
		Consequent{Variable: "speed", Term: "fast"},
	)

	c1 := r.Conditions()
	c2 := r.Conditions()

	changed := "changed"
	c1[0].Variable = changed

	if c2[0].Variable == changed {
		t.Fatal("expected defensive copy from Conditions()")
	}
}

func TestCondition_struct(t *testing.T) {
	t.Parallel()

	c := Condition{Variable: "temp", Term: "hot"}

	if c.Variable != "temp" {
		t.Fatalf("expected temp, got %s", c.Variable)
	}

	if c.Term != "hot" {
		t.Fatalf("expected hot, got %s", c.Term)
	}
}

func TestConsequent_struct(t *testing.T) {
	t.Parallel()

	c := Consequent{Variable: "speed", Term: "fast"}

	if c.Variable != "speed" {
		t.Fatalf("expected speed, got %s", c.Variable)
	}

	if c.Term != "fast" {
		t.Fatalf("expected fast, got %s", c.Term)
	}
}

func TestOperator_constants(t *testing.T) {
	t.Parallel()

	if And != 0 {
		t.Fatalf("expected And=0, got %d", And)
	}

	if Or != 1 {
		t.Fatalf("expected Or=1, got %d", Or)
	}
}
