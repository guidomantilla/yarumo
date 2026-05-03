package constraints

import "testing"

func signedIdentity[T Signed](v T) T { return v }

func unsignedIdentity[T Unsigned](v T) T { return v }

func integerIdentity[T Integer](v T) T { return v }

func floatIdentity[T Float](v T) T { return v }

func complexIdentity[T Complex](v T) T { return v }

func numberIdentity[T Number](v T) T { return v }

func ordenableIdentity[T Ordenable](v T) T { return v }

func comparableIdentity[T Comparable](v T) T { return v }

func TestSigned(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		if signedIdentity[int](-1) != -1 {
			t.Fatal("expected -1")
		}
	})

	t.Run("int8", func(t *testing.T) {
		t.Parallel()

		if signedIdentity[int8](-8) != -8 {
			t.Fatal("expected -8")
		}
	})

	t.Run("int16", func(t *testing.T) {
		t.Parallel()

		if signedIdentity[int16](-16) != -16 {
			t.Fatal("expected -16")
		}
	})

	t.Run("int32", func(t *testing.T) {
		t.Parallel()

		if signedIdentity[int32](-32) != -32 {
			t.Fatal("expected -32")
		}
	})

	t.Run("int64", func(t *testing.T) {
		t.Parallel()

		if signedIdentity[int64](-64) != -64 {
			t.Fatal("expected -64")
		}
	})
}

func TestUnsigned(t *testing.T) {
	t.Parallel()

	t.Run("uint", func(t *testing.T) {
		t.Parallel()

		if unsignedIdentity[uint](1) != 1 {
			t.Fatal("expected 1")
		}
	})

	t.Run("uint8", func(t *testing.T) {
		t.Parallel()

		if unsignedIdentity[uint8](8) != 8 {
			t.Fatal("expected 8")
		}
	})

	t.Run("uint16", func(t *testing.T) {
		t.Parallel()

		if unsignedIdentity[uint16](16) != 16 {
			t.Fatal("expected 16")
		}
	})

	t.Run("uint32", func(t *testing.T) {
		t.Parallel()

		if unsignedIdentity[uint32](32) != 32 {
			t.Fatal("expected 32")
		}
	})

	t.Run("uint64", func(t *testing.T) {
		t.Parallel()

		if unsignedIdentity[uint64](64) != 64 {
			t.Fatal("expected 64")
		}
	})

	t.Run("uintptr", func(t *testing.T) {
		t.Parallel()

		if unsignedIdentity[uintptr](42) != 42 {
			t.Fatal("expected 42")
		}
	})
}

func TestInteger(t *testing.T) {
	t.Parallel()

	t.Run("signed int", func(t *testing.T) {
		t.Parallel()

		if integerIdentity[int](-5) != -5 {
			t.Fatal("expected -5")
		}
	})

	t.Run("unsigned uint", func(t *testing.T) {
		t.Parallel()

		if integerIdentity[uint](5) != 5 {
			t.Fatal("expected 5")
		}
	})
}

func TestFloat(t *testing.T) {
	t.Parallel()

	t.Run("float32", func(t *testing.T) {
		t.Parallel()

		if floatIdentity[float32](3.14) != 3.14 {
			t.Fatal("expected 3.14")
		}
	})

	t.Run("float64", func(t *testing.T) {
		t.Parallel()

		if floatIdentity[float64](2.718) != 2.718 {
			t.Fatal("expected 2.718")
		}
	})
}

func TestComplex(t *testing.T) {
	t.Parallel()

	t.Run("complex64", func(t *testing.T) {
		t.Parallel()

		if complexIdentity[complex64](1+2i) != 1+2i {
			t.Fatal("expected 1+2i")
		}
	})

	t.Run("complex128", func(t *testing.T) {
		t.Parallel()

		if complexIdentity[complex128](3+4i) != 3+4i {
			t.Fatal("expected 3+4i")
		}
	})
}

func TestNumber(t *testing.T) {
	t.Parallel()

	t.Run("integer type", func(t *testing.T) {
		t.Parallel()

		if numberIdentity[int](10) != 10 {
			t.Fatal("expected 10")
		}
	})

	t.Run("float type", func(t *testing.T) {
		t.Parallel()

		if numberIdentity[float64](1.5) != 1.5 {
			t.Fatal("expected 1.5")
		}
	})
}

func TestComparable(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		if comparableIdentity[int](7) != 7 {
			t.Fatal("expected 7")
		}
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		if comparableIdentity[string]("abc") != "abc" {
			t.Fatal("expected abc")
		}
	})
}

func TestOrdenable(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		if ordenableIdentity[int](3) != 3 {
			t.Fatal("expected 3")
		}
	})

	t.Run("float64", func(t *testing.T) {
		t.Parallel()

		if ordenableIdentity[float64](2.5) != 2.5 {
			t.Fatal("expected 2.5")
		}
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		if ordenableIdentity[string]("z") != "z" {
			t.Fatal("expected z")
		}
	})
}

// Derived types verify that the ~ operator accepts user-defined types.
type myInt int

type myFloat float64

type myUint uint

func TestSigned_DerivedType(t *testing.T) {
	t.Parallel()

	if signedIdentity[myInt](42) != 42 {
		t.Fatal("expected 42")
	}
}

func TestUnsigned_DerivedType(t *testing.T) {
	t.Parallel()

	if unsignedIdentity[myUint](7) != 7 {
		t.Fatal("expected 7")
	}
}

func TestFloat_DerivedType(t *testing.T) {
	t.Parallel()

	if floatIdentity[myFloat](1.1) != 1.1 {
		t.Fatal("expected 1.1")
	}
}

func TestNumber_DerivedType(t *testing.T) {
	t.Parallel()

	if numberIdentity[myInt](99) != 99 {
		t.Fatal("expected 99")
	}
}

func TestInteger_DerivedType(t *testing.T) {
	t.Parallel()

	if integerIdentity[myInt](10) != 10 {
		t.Fatal("expected 10")
	}
}
