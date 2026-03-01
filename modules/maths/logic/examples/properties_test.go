package examples

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"
)

// TestProperties_foundational verifies fundamental laws of propositional logic.
func TestProperties_foundational(t *testing.T) {
	t.Parallel()

	p := logic.Var("P")
	q := logic.Var("Q")
	r := logic.Var("R")

	t.Run("double negation", func(t *testing.T) {
		t.Parallel()

		if !logic.Equivalent(logic.NotF{F: logic.NotF{F: p}}, p) {
			t.Fatal("law failed")
		}
	})

	t.Run("idempotent and", func(t *testing.T) {
		t.Parallel()

		if !logic.Equivalent(logic.AndF{L: p, R: p}, p) {
			t.Fatal("law failed")
		}
	})

	t.Run("idempotent or", func(t *testing.T) {
		t.Parallel()

		if !logic.Equivalent(logic.OrF{L: p, R: p}, p) {
			t.Fatal("law failed")
		}
	})

	t.Run("commutative and", func(t *testing.T) {
		t.Parallel()

		left := logic.AndF{L: p, R: q}
		right := logic.AndF{L: q, R: p}

		if !logic.Equivalent(left, right) {
			t.Fatal("law failed")
		}
	})

	t.Run("commutative or", func(t *testing.T) {
		t.Parallel()

		left := logic.OrF{L: p, R: q}
		right := logic.OrF{L: q, R: p}

		if !logic.Equivalent(left, right) {
			t.Fatal("law failed")
		}
	})

	t.Run("associative and", func(t *testing.T) {
		t.Parallel()

		left := logic.AndF{L: logic.AndF{L: p, R: q}, R: r}
		right := logic.AndF{L: p, R: logic.AndF{L: q, R: r}}

		if !logic.Equivalent(left, right) {
			t.Fatal("law failed")
		}
	})

	t.Run("associative or", func(t *testing.T) {
		t.Parallel()

		left := logic.OrF{L: logic.OrF{L: p, R: q}, R: r}
		right := logic.OrF{L: p, R: logic.OrF{L: q, R: r}}

		if !logic.Equivalent(left, right) {
			t.Fatal("law failed")
		}
	})

	t.Run("distributive and over or", func(t *testing.T) {
		t.Parallel()

		left := logic.AndF{L: p, R: logic.OrF{L: q, R: r}}
		right := logic.OrF{
			L: logic.AndF{L: p, R: q},
			R: logic.AndF{L: p, R: r},
		}

		if !logic.Equivalent(left, right) {
			t.Fatal("law failed")
		}
	})

	t.Run("distributive or over and", func(t *testing.T) {
		t.Parallel()

		left := logic.OrF{L: p, R: logic.AndF{L: q, R: r}}
		right := logic.AndF{
			L: logic.OrF{L: p, R: q},
			R: logic.OrF{L: p, R: r},
		}

		if !logic.Equivalent(left, right) {
			t.Fatal("law failed")
		}
	})

	t.Run("de morgan and", func(t *testing.T) {
		t.Parallel()

		left := logic.NotF{F: logic.AndF{L: p, R: q}}
		right := logic.OrF{L: logic.NotF{F: p}, R: logic.NotF{F: q}}

		if !logic.Equivalent(left, right) {
			t.Fatal("law failed")
		}
	})

	t.Run("de morgan or", func(t *testing.T) {
		t.Parallel()

		left := logic.NotF{F: logic.OrF{L: p, R: q}}
		right := logic.AndF{L: logic.NotF{F: p}, R: logic.NotF{F: q}}

		if !logic.Equivalent(left, right) {
			t.Fatal("law failed")
		}
	})
}

// TestProperties_identity verifies identity, complement, absorption, and compound laws.
func TestProperties_identity(t *testing.T) {
	t.Parallel()

	p := logic.Var("P")
	q := logic.Var("Q")

	t.Run("identity and", func(t *testing.T) {
		t.Parallel()

		if !logic.Equivalent(logic.AndF{L: p, R: logic.TrueF{}}, p) {
			t.Fatal("law failed")
		}
	})

	t.Run("identity or", func(t *testing.T) {
		t.Parallel()

		if !logic.Equivalent(logic.OrF{L: p, R: logic.FalseF{}}, p) {
			t.Fatal("law failed")
		}
	})

	t.Run("complement and", func(t *testing.T) {
		t.Parallel()

		if !logic.Equivalent(logic.AndF{L: p, R: logic.NotF{F: p}}, logic.FalseF{}) {
			t.Fatal("law failed")
		}
	})

	t.Run("complement or", func(t *testing.T) {
		t.Parallel()

		if !logic.Equivalent(logic.OrF{L: p, R: logic.NotF{F: p}}, logic.TrueF{}) {
			t.Fatal("law failed")
		}
	})

	t.Run("annihilation and", func(t *testing.T) {
		t.Parallel()

		if !logic.Equivalent(logic.AndF{L: p, R: logic.FalseF{}}, logic.FalseF{}) {
			t.Fatal("law failed")
		}
	})

	t.Run("annihilation or", func(t *testing.T) {
		t.Parallel()

		if !logic.Equivalent(logic.OrF{L: p, R: logic.TrueF{}}, logic.TrueF{}) {
			t.Fatal("law failed")
		}
	})

	t.Run("absorption and", func(t *testing.T) {
		t.Parallel()

		left := logic.AndF{L: p, R: logic.OrF{L: p, R: q}}

		if !logic.Equivalent(left, p) {
			t.Fatal("law failed")
		}
	})

	t.Run("absorption or", func(t *testing.T) {
		t.Parallel()

		left := logic.OrF{L: p, R: logic.AndF{L: p, R: q}}

		if !logic.Equivalent(left, p) {
			t.Fatal("law failed")
		}
	})

	t.Run("implication elimination", func(t *testing.T) {
		t.Parallel()

		left := logic.ImplF{L: p, R: q}
		right := logic.OrF{L: logic.NotF{F: p}, R: q}

		if !logic.Equivalent(left, right) {
			t.Fatal("law failed")
		}
	})

	t.Run("biconditional elimination", func(t *testing.T) {
		t.Parallel()

		left := logic.IffF{L: p, R: q}
		right := logic.AndF{
			L: logic.ImplF{L: p, R: q},
			R: logic.ImplF{L: q, R: p},
		}

		if !logic.Equivalent(left, right) {
			t.Fatal("law failed")
		}
	})

	t.Run("contrapositive", func(t *testing.T) {
		t.Parallel()

		left := logic.ImplF{L: p, R: q}
		right := logic.ImplF{L: logic.NotF{F: q}, R: logic.NotF{F: p}}

		if !logic.Equivalent(left, right) {
			t.Fatal("law failed")
		}
	})
}
