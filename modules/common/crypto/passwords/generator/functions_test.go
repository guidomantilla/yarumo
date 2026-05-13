package generator

import (
	"errors"
	"slices"
	"testing"

	crandom "github.com/guidomantilla/yarumo/common/random"
)

// withRandomOverrides swaps every package-level CSPRNG indirection for the
// duration of fn. It restores the originals via defer; tests using it
// cannot be parallel (they mutate package state).
func withRandomOverrides(special, number, upper, lower crandom.TextFn, num crandom.NumberFn, fn func()) {
	origSpecial, origNumber, origUpper, origLower, origNum := textSpecial, textNumber, textUpper, textLower, randNumber
	if special != nil {
		textSpecial = special
	}
	if number != nil {
		textNumber = number
	}
	if upper != nil {
		textUpper = upper
	}
	if lower != nil {
		textLower = lower
	}
	if num != nil {
		randNumber = num
	}
	defer func() {
		textSpecial, textNumber, textUpper, textLower, randNumber = origSpecial, origNumber, origUpper, origLower, origNum
	}()
	fn()
}

func errFn(_ int) (string, error) { return "", errors.New("boom") }

func TestGenerate_NilReceiver(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrGeneratorIsNil", func(t *testing.T) {
		t.Parallel()

		_, err := generate(nil)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrGeneratorIsNil) {
			t.Fatalf("expected ErrGeneratorIsNil, got %v", err)
		}
		if !errors.Is(err, ErrGenerationFailed) {
			t.Fatalf("expected ErrGenerationFailed in chain, got %v", err)
		}
	})
}

func TestValidate_NilReceiver(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrGeneratorIsNil", func(t *testing.T) {
		t.Parallel()

		err := validate(nil, "whatever")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrGeneratorIsNil) {
			t.Fatalf("expected ErrGeneratorIsNil, got %v", err)
		}
		if !errors.Is(err, ErrValidationFailed) {
			t.Fatalf("expected ErrValidationFailed in chain, got %v", err)
		}
	})
}

func TestGenerate_SpecialError(t *testing.T) { //nolint:paralleltest
	withRandomOverrides(errFn, nil, nil, nil, nil, func() {
		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		_, errGen := g.Generate()
		if errGen == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(errGen, ErrGenerationFailed) {
			t.Fatalf("expected ErrGenerationFailed, got %v", errGen)
		}
	})
}

func TestGenerate_NumberError(t *testing.T) { //nolint:paralleltest
	withRandomOverrides(nil, errFn, nil, nil, nil, func() {
		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		_, errGen := g.Generate()
		if errGen == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(errGen, ErrGenerationFailed) {
			t.Fatalf("expected ErrGenerationFailed, got %v", errGen)
		}
	})
}

func TestGenerate_UpperError(t *testing.T) { //nolint:paralleltest
	withRandomOverrides(nil, nil, errFn, nil, nil, func() {
		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		_, errGen := g.Generate()
		if errGen == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(errGen, ErrGenerationFailed) {
			t.Fatalf("expected ErrGenerationFailed, got %v", errGen)
		}
	})
}

func TestGenerate_LowerError(t *testing.T) { //nolint:paralleltest
	// The minimum-lowercase call comes before the fill-lowercase call;
	// either one returning error suffices for this path.
	withRandomOverrides(nil, nil, nil, errFn, nil, func() {
		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		_, errGen := g.Generate()
		if errGen == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(errGen, ErrGenerationFailed) {
			t.Fatalf("expected ErrGenerationFailed, got %v", errGen)
		}
	})
}

func TestGenerate_FillError(t *testing.T) { //nolint:paralleltest
	// Make textLower succeed once (for minLowerCase) and fail on the
	// second call (the fill). Use a counter.
	var calls int
	fillFn := func(size int) (string, error) {
		calls++
		if calls == 1 {
			return crandom.TextLower(size)
		}
		return "", errors.New("fill boom")
	}
	withRandomOverrides(nil, nil, nil, fillFn, nil, func() {
		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		_, errGen := g.Generate()
		if errGen == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(errGen, ErrGenerationFailed) {
			t.Fatalf("expected ErrGenerationFailed, got %v", errGen)
		}
	})
}

func TestGenerate_ShuffleError(t *testing.T) { //nolint:paralleltest
	withRandomOverrides(nil, nil, nil, nil, func(int64) (int64, error) {
		return 0, errors.New("shuffle boom")
	}, func() {
		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		_, errGen := g.Generate()
		if errGen == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(errGen, ErrShuffleFailed) {
			t.Fatalf("expected ErrShuffleFailed, got %v", errGen)
		}
	})
}

func TestShuffleRunes_Error(t *testing.T) { //nolint:paralleltest
	withRandomOverrides(nil, nil, nil, nil, func(int64) (int64, error) {
		return 0, errors.New("boom")
	}, func() {
		runes := []rune("abcd")
		err := shuffleRunes(runes)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrShuffleFailed) {
			t.Fatalf("expected ErrShuffleFailed, got %v", err)
		}
	})
}

func TestShuffleRunes(t *testing.T) {
	t.Parallel()

	t.Run("preserves multiset of runes", func(t *testing.T) {
		t.Parallel()

		original := []rune("abcdefghijABCDEFGHIJ0123456789@#$%")
		copyOf := make([]rune, len(original))
		copy(copyOf, original)

		err := shuffleRunes(copyOf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Sort and compare to confirm same multiset.
		sortedA := make([]rune, len(original))
		copy(sortedA, original)
		slices.Sort(sortedA)

		sortedB := make([]rune, len(copyOf))
		copy(sortedB, copyOf)
		slices.Sort(sortedB)

		if string(sortedA) != string(sortedB) {
			t.Fatalf("multiset changed: before sorted=%q after sorted=%q", string(sortedA), string(sortedB))
		}
	})

	t.Run("empty slice is a no-op", func(t *testing.T) {
		t.Parallel()

		var empty []rune
		err := shuffleRunes(empty)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(empty) != 0 {
			t.Fatalf("expected empty, got len %d", len(empty))
		}
	})

	t.Run("single element slice is a no-op", func(t *testing.T) {
		t.Parallel()

		one := []rune{'x'}
		err := shuffleRunes(one)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(one) != 1 || one[0] != 'x' {
			t.Fatalf("expected [x], got %q", string(one))
		}
	})
}

// FuzzGenerateValidate is the Generate -> Validate round-trip fuzz target.
//
// The fuzzer drives the constraint space (length, minimums); for every
// internally-consistent configuration, a freshly generated password must
// satisfy its own validator. A panic or validation failure on a generated
// password indicates a regression in either the generator or the validator.
func FuzzGenerateValidate(f *testing.F) {
	// Seed corpus covers default + tight + zero-everything configurations.
	f.Add(26, 4, 6, 6, 6)
	f.Add(8, 2, 2, 2, 2)
	f.Add(0, 0, 0, 0, 0)
	f.Add(40, 0, 0, 0, 40)
	f.Add(32, 8, 8, 8, 8)

	f.Fuzz(func(t *testing.T, length, special, number, upper, lower int) {
		// Bound the inputs so the test does not allocate gigabytes
		// (length >= 0 already implied by clamping in With<Field>).
		if length < 0 || length > 512 {
			t.Skip()
		}
		if special < 0 || number < 0 || upper < 0 || lower < 0 {
			t.Skip()
		}
		if special > 512 || number > 512 || upper > 512 || lower > 512 {
			t.Skip()
		}

		g, err := NewGenerator(
			WithPasswordLength(length),
			WithMinSpecialChar(special),
			WithMinNumber(number),
			WithMinUpperCase(upper),
			WithMinLowerCase(lower),
		)
		if err != nil {
			// Constraint-violating configs are expected to fail; they're not interesting.
			return
		}

		pw, err := g.Generate()
		if err != nil {
			t.Fatalf("Generate failed for valid config len=%d special=%d num=%d up=%d low=%d: %v",
				length, special, number, upper, lower, err)
		}

		if len(pw) != length {
			t.Fatalf("length mismatch: got %d, want %d", len(pw), length)
		}

		errValidate := g.Validate(pw)
		if errValidate != nil {
			t.Fatalf("generated password failed its own validator: %v (pw=%q config: len=%d special=%d num=%d up=%d low=%d)",
				errValidate, pw, length, special, number, upper, lower)
		}
	})
}
