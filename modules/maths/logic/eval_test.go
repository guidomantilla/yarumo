package logic

import "testing"

func TestVar_Eval(t *testing.T) {
	t.Parallel()

	t.Run("present and true", func(t *testing.T) {
		t.Parallel()

		got := Var("A").Eval(Fact{"A": true})
		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("present and false", func(t *testing.T) {
		t.Parallel()

		got := Var("A").Eval(Fact{"A": false})
		if got {
			t.Fatal("expected false")
		}
	})

	t.Run("missing defaults to false", func(t *testing.T) {
		t.Parallel()

		got := Var("A").Eval(Fact{})
		if got {
			t.Fatal("expected false for missing variable")
		}
	})

	t.Run("nil facts defaults to false", func(t *testing.T) {
		t.Parallel()

		got := Var("A").Eval(nil)
		if got {
			t.Fatal("expected false for nil facts")
		}
	})
}

func TestTrueF_Eval(t *testing.T) {
	t.Parallel()

	t.Run("always true with empty facts", func(t *testing.T) {
		t.Parallel()

		got := TrueF{}.Eval(Fact{})
		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("always true with nil facts", func(t *testing.T) {
		t.Parallel()

		got := TrueF{}.Eval(nil)
		if !got {
			t.Fatal("expected true")
		}
	})
}

func TestFalseF_Eval(t *testing.T) {
	t.Parallel()

	t.Run("always false with empty facts", func(t *testing.T) {
		t.Parallel()

		got := FalseF{}.Eval(Fact{})
		if got {
			t.Fatal("expected false")
		}
	})

	t.Run("always false with nil facts", func(t *testing.T) {
		t.Parallel()

		got := FalseF{}.Eval(nil)
		if got {
			t.Fatal("expected false")
		}
	})
}

func TestNotF_Eval(t *testing.T) {
	t.Parallel()

	t.Run("negates true", func(t *testing.T) {
		t.Parallel()

		got := NotF{F: Var("A")}.Eval(Fact{"A": true})
		if got {
			t.Fatal("expected false")
		}
	})

	t.Run("negates false", func(t *testing.T) {
		t.Parallel()

		got := NotF{F: Var("A")}.Eval(Fact{"A": false})
		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("double negation", func(t *testing.T) {
		t.Parallel()

		got := NotF{F: NotF{F: Var("A")}}.Eval(Fact{"A": true})
		if !got {
			t.Fatal("expected true")
		}
	})
}

func TestAndF_Eval(t *testing.T) {
	t.Parallel()

	t.Run("true and true", func(t *testing.T) {
		t.Parallel()

		got := AndF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": true, "B": true})
		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("true and false", func(t *testing.T) {
		t.Parallel()

		got := AndF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": true, "B": false})
		if got {
			t.Fatal("expected false")
		}
	})

	t.Run("false and true", func(t *testing.T) {
		t.Parallel()

		got := AndF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": false, "B": true})
		if got {
			t.Fatal("expected false")
		}
	})

	t.Run("false and false", func(t *testing.T) {
		t.Parallel()

		got := AndF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": false, "B": false})
		if got {
			t.Fatal("expected false")
		}
	})
}

func TestOrF_Eval(t *testing.T) {
	t.Parallel()

	t.Run("true or true", func(t *testing.T) {
		t.Parallel()

		got := OrF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": true, "B": true})
		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("true or false", func(t *testing.T) {
		t.Parallel()

		got := OrF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": true, "B": false})
		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("false or true", func(t *testing.T) {
		t.Parallel()

		got := OrF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": false, "B": true})
		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("false or false", func(t *testing.T) {
		t.Parallel()

		got := OrF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": false, "B": false})
		if got {
			t.Fatal("expected false")
		}
	})
}

func TestImplF_Eval(t *testing.T) {
	t.Parallel()

	t.Run("true implies true", func(t *testing.T) {
		t.Parallel()

		got := ImplF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": true, "B": true})
		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("true implies false", func(t *testing.T) {
		t.Parallel()

		got := ImplF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": true, "B": false})
		if got {
			t.Fatal("expected false")
		}
	})

	t.Run("false implies true", func(t *testing.T) {
		t.Parallel()

		got := ImplF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": false, "B": true})
		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("false implies false", func(t *testing.T) {
		t.Parallel()

		got := ImplF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": false, "B": false})
		if !got {
			t.Fatal("expected true")
		}
	})
}

func TestIffF_Eval(t *testing.T) {
	t.Parallel()

	t.Run("true iff true", func(t *testing.T) {
		t.Parallel()

		got := IffF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": true, "B": true})
		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("true iff false", func(t *testing.T) {
		t.Parallel()

		got := IffF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": true, "B": false})
		if got {
			t.Fatal("expected false")
		}
	})

	t.Run("false iff true", func(t *testing.T) {
		t.Parallel()

		got := IffF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": false, "B": true})
		if got {
			t.Fatal("expected false")
		}
	})

	t.Run("false iff false", func(t *testing.T) {
		t.Parallel()

		got := IffF{L: Var("A"), R: Var("B")}.Eval(Fact{"A": false, "B": false})
		if !got {
			t.Fatal("expected true")
		}
	})
}

func TestGroupF_Eval(t *testing.T) {
	t.Parallel()

	t.Run("delegates to inner formula", func(t *testing.T) {
		t.Parallel()

		got := GroupF{F: Var("A")}.Eval(Fact{"A": true})
		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("delegates false", func(t *testing.T) {
		t.Parallel()

		got := GroupF{F: Var("A")}.Eval(Fact{"A": false})
		if got {
			t.Fatal("expected false")
		}
	})
}

func TestEval_nested(t *testing.T) {
	t.Parallel()

	t.Run("modus ponens", func(t *testing.T) {
		t.Parallel()
		// (A & (A => B)) should give true when A=true, B=true
		f := AndF{L: Var("A"), R: ImplF{L: Var("A"), R: Var("B")}}

		got := f.Eval(Fact{"A": true, "B": true})
		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("de morgan", func(t *testing.T) {
		t.Parallel()
		// !(A | B) should equal (!A & !B) for A=false, B=false
		left := NotF{F: OrF{L: Var("A"), R: Var("B")}}
		right := AndF{L: NotF{F: Var("A")}, R: NotF{F: Var("B")}}
		facts := Fact{"A": false, "B": false}
		l := left.Eval(facts)

		r := right.Eval(facts)
		if l != r {
			t.Fatalf("expected equal, got left=%v right=%v", l, r)
		}
	})
}
