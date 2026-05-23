package validation

import (
	"errors"
	"strings"
	"testing"
)

func TestIsRequired(t *testing.T) {
	t.Parallel()

	t.Run("happy path string", func(t *testing.T) {
		t.Parallel()

		err := IsRequired("hello")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path int", func(t *testing.T) {
		t.Parallel()

		err := IsRequired(42)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty string", func(t *testing.T) {
		t.Parallel()

		err := IsRequired("")
		if !errors.Is(err, ErrFieldRequired) {
			t.Fatalf("expected ErrFieldRequired, got %v", err)
		}
	})

	t.Run("error zero int", func(t *testing.T) {
		t.Parallel()

		err := IsRequired(0)
		if !errors.Is(err, ErrFieldRequired) {
			t.Fatalf("expected ErrFieldRequired, got %v", err)
		}
	})
}

func TestMustBeUndefined(t *testing.T) {
	t.Parallel()

	t.Run("happy path empty string", func(t *testing.T) {
		t.Parallel()

		err := MustBeUndefined("")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error non-empty string", func(t *testing.T) {
		t.Parallel()

		err := MustBeUndefined("hello")
		if !errors.Is(err, ErrFieldMustBeUndefined) {
			t.Fatalf("expected ErrFieldMustBeUndefined, got %v", err)
		}
	})

	t.Run("happy path zero int", func(t *testing.T) {
		t.Parallel()

		err := MustBeUndefined(0)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestMinLen(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := MinLen("hello", 3)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error below", func(t *testing.T) {
		t.Parallel()

		err := MinLen("hi", 5)
		if !errors.Is(err, ErrMinLen) {
			t.Fatalf("expected ErrMinLen, got %v", err)
		}
	})

	t.Run("negative threshold accepts empty", func(t *testing.T) {
		t.Parallel()

		err := MinLen("", -3)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestMaxLen(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := MaxLen("hello", 10)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error above", func(t *testing.T) {
		t.Parallel()

		err := MaxLen("hello world", 5)
		if !errors.Is(err, ErrMaxLen) {
			t.Fatalf("expected ErrMaxLen, got %v", err)
		}
	})

	t.Run("negative threshold rejects non-empty", func(t *testing.T) {
		t.Parallel()

		err := MaxLen("x", -1)
		if !errors.Is(err, ErrMaxLen) {
			t.Fatalf("expected ErrMaxLen, got %v", err)
		}
	})
}

func TestMatchesRegex(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := MatchesRegex("abc123", `^[a-z]+\d+$`)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error mismatch", func(t *testing.T) {
		t.Parallel()

		err := MatchesRegex("ABC", `^[a-z]+$`)
		if !errors.Is(err, ErrRegexMismatch) {
			t.Fatalf("expected ErrRegexMismatch, got %v", err)
		}
	})

	t.Run("error invalid pattern", func(t *testing.T) {
		t.Parallel()

		err := MatchesRegex("x", `[`)
		if !errors.Is(err, ErrRegexInvalid) {
			t.Fatalf("expected ErrRegexInvalid, got %v", err)
		}
	})
}

func TestIsEmail(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsEmail("a@b.com")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsEmail("")
		if !errors.Is(err, ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid, got %v", err)
		}
	})

	t.Run("error malformed", func(t *testing.T) {
		t.Parallel()

		err := IsEmail("not-an-email")
		if !errors.Is(err, ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid, got %v", err)
		}
	})

	t.Run("error with display name", func(t *testing.T) {
		t.Parallel()

		err := IsEmail("Name <a@b.com>")
		if !errors.Is(err, ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid, got %v", err)
		}
	})
}

func TestIsURL(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsURL("https://example.com/path")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsURL("")
		if !errors.Is(err, ErrURLInvalid) {
			t.Fatalf("expected ErrURLInvalid, got %v", err)
		}
	})

	t.Run("error missing scheme", func(t *testing.T) {
		t.Parallel()

		err := IsURL("example.com")
		if !errors.Is(err, ErrURLInvalid) {
			t.Fatalf("expected ErrURLInvalid, got %v", err)
		}
	})

	t.Run("error unparseable", func(t *testing.T) {
		t.Parallel()

		err := IsURL("http://[::1")
		if !errors.Is(err, ErrURLInvalid) {
			t.Fatalf("expected ErrURLInvalid, got %v", err)
		}
	})
}

func TestMin(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := Min(10, 5)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error below", func(t *testing.T) {
		t.Parallel()

		err := Min(3, 10)
		if !errors.Is(err, ErrMinValue) {
			t.Fatalf("expected ErrMinValue, got %v", err)
		}
	})

	t.Run("happy path float", func(t *testing.T) {
		t.Parallel()

		err := Min(1.5, 1.0)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestMax(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := Max(3, 10)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error above", func(t *testing.T) {
		t.Parallel()

		err := Max(20, 10)
		if !errors.Is(err, ErrMaxValue) {
			t.Fatalf("expected ErrMaxValue, got %v", err)
		}
	})
}

func TestInRange(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := InRange(5, 0, 10)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error below", func(t *testing.T) {
		t.Parallel()

		err := InRange(-1, 0, 10)
		if !errors.Is(err, ErrOutOfRange) {
			t.Fatalf("expected ErrOutOfRange, got %v", err)
		}
	})

	t.Run("error above", func(t *testing.T) {
		t.Parallel()

		err := InRange(20, 0, 10)
		if !errors.Is(err, ErrOutOfRange) {
			t.Fatalf("expected ErrOutOfRange, got %v", err)
		}
	})

	t.Run("error invalid range", func(t *testing.T) {
		t.Parallel()

		err := InRange(5, 10, 0)
		if !errors.Is(err, ErrInvalidRange) {
			t.Fatalf("expected ErrInvalidRange, got %v", err)
		}
	})
}

func TestIsUUID(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsUUID("550e8400-e29b-41d4-a716-446655440000")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsUUID("")
		if !errors.Is(err, ErrUUIDInvalid) {
			t.Fatalf("expected ErrUUIDInvalid, got %v", err)
		}
	})

	t.Run("error malformed", func(t *testing.T) {
		t.Parallel()

		err := IsUUID("not-a-uuid")
		if !errors.Is(err, ErrUUIDInvalid) {
			t.Fatalf("expected ErrUUIDInvalid, got %v", err)
		}
	})
}

func TestIsULID(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsULID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsULID("")
		if !errors.Is(err, ErrULIDInvalid) {
			t.Fatalf("expected ErrULIDInvalid, got %v", err)
		}
	})

	t.Run("error malformed", func(t *testing.T) {
		t.Parallel()

		err := IsULID("not-a-ulid")
		if !errors.Is(err, ErrULIDInvalid) {
			t.Fatalf("expected ErrULIDInvalid, got %v", err)
		}
	})
}

func TestNonEmpty(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := NonEmpty([]int{1, 2, 3})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := NonEmpty([]int{})
		if !errors.Is(err, ErrCollectionEmpty) {
			t.Fatalf("expected ErrCollectionEmpty, got %v", err)
		}
	})

	t.Run("error nil slice", func(t *testing.T) {
		t.Parallel()

		var xs []string

		err := NonEmpty(xs)
		if !errors.Is(err, ErrCollectionEmpty) {
			t.Fatalf("expected ErrCollectionEmpty, got %v", err)
		}
	})
}

func TestEach(t *testing.T) {
	t.Parallel()

	t.Run("happy path all pass", func(t *testing.T) {
		t.Parallel()

		err := Each([]string{"a@b.com", "c@d.com"}, IsEmail)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error one fails", func(t *testing.T) {
		t.Parallel()

		err := Each([]string{"a@b.com", "not-an-email"}, IsEmail)
		if !errors.Is(err, ErrEachFailed) {
			t.Fatalf("expected ErrEachFailed, got %v", err)
		}

		if !errors.Is(err, ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid to be wrapped, got %v", err)
		}
	})

	t.Run("nil check no-ops", func(t *testing.T) {
		t.Parallel()

		err := Each([]int{1, 2, 3}, nil)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("empty slice trivially passes", func(t *testing.T) {
		t.Parallel()

		err := Each([]int{}, func(int) error { return errors.New("never called") })
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

type owner struct {
	Email string
	Tags  []string
}

type pokemon struct {
	Name  string
	Owner owner
	IDs   []int
}

func TestGetField(t *testing.T) {
	t.Parallel()

	t.Run("dotted path struct", func(t *testing.T) {
		t.Parallel()

		p := pokemon{Name: "pikachu", Owner: owner{Email: "ash@kanto.com"}}

		v, err := GetField(p, "Owner.Email")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := v.(string)
		if !ok || got != "ash@kanto.com" {
			t.Fatalf("expected ash@kanto.com, got %v", v)
		}
	})

	t.Run("slice index", func(t *testing.T) {
		t.Parallel()

		p := pokemon{IDs: []int{10, 20, 30}}

		v, err := GetField(p, "IDs[1]")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := v.(int)
		if !ok || got != 20 {
			t.Fatalf("expected 20, got %v", v)
		}
	})

	t.Run("slice index nested", func(t *testing.T) {
		t.Parallel()

		p := pokemon{Owner: owner{Tags: []string{"trainer", "champion"}}}

		v, err := GetField(p, "Owner.Tags[0]")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := v.(string)
		if !ok || got != "trainer" {
			t.Fatalf("expected trainer, got %v", v)
		}
	})

	t.Run("pointer auto-deref", func(t *testing.T) {
		t.Parallel()

		p := &pokemon{Name: "snorlax"}

		v, err := GetField(p, "Name")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := v.(string)
		if !ok || got != "snorlax" {
			t.Fatalf("expected snorlax, got %v", v)
		}
	})

	t.Run("map lookup", func(t *testing.T) {
		t.Parallel()

		m := map[string]any{"key": "value"}

		v, err := GetField(m, "key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := v.(string)
		if !ok || got != "value" {
			t.Fatalf("expected value, got %v", v)
		}
	})

	t.Run("error nil object", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(nil, "X")
		if !errors.Is(err, ErrObjectNil) {
			t.Fatalf("expected ErrObjectNil, got %v", err)
		}
	})

	t.Run("error empty path", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error missing field", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "Nope")
		if !errors.Is(err, ErrPathNotFound) {
			t.Fatalf("expected ErrPathNotFound, got %v", err)
		}
	})

	t.Run("error type mismatch", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{Name: "x"}, "Name.Inner")
		if !errors.Is(err, ErrPathTypeMismatch) {
			t.Fatalf("expected ErrPathTypeMismatch, got %v", err)
		}
	})

	t.Run("error index out of range", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{IDs: []int{1}}, "IDs[5]")
		if !errors.Is(err, ErrIndexOutOfRange) {
			t.Fatalf("expected ErrIndexOutOfRange, got %v", err)
		}
	})

	t.Run("error index on non-slice", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{Name: "x"}, "Name[0]")
		if !errors.Is(err, ErrPathTypeMismatch) {
			t.Fatalf("expected ErrPathTypeMismatch, got %v", err)
		}
	})

	t.Run("error malformed bracket", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "IDs[")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error empty index", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "IDs[]")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error non-numeric index", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "IDs[abc]")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error negative index", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "IDs[-1]")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error leading bracket", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "[0]")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error nil pointer in path", func(t *testing.T) {
		t.Parallel()

		type wrapper struct {
			P *pokemon
		}

		_, err := GetField(wrapper{}, "P.Name")
		if !errors.Is(err, ErrPathNotFound) {
			t.Fatalf("expected ErrPathNotFound, got %v", err)
		}
	})

	t.Run("error map non-string key", func(t *testing.T) {
		t.Parallel()

		m := map[int]string{1: "x"}

		_, err := GetField(m, "1")
		if !errors.Is(err, ErrPathTypeMismatch) {
			t.Fatalf("expected ErrPathTypeMismatch, got %v", err)
		}
	})

	t.Run("error map missing key", func(t *testing.T) {
		t.Parallel()

		m := map[string]any{"a": 1}

		_, err := GetField(m, "missing")
		if !errors.Is(err, ErrPathNotFound) {
			t.Fatalf("expected ErrPathNotFound, got %v", err)
		}
	})

	t.Run("error trailing chars after bracket", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "IDs[0]junk")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error empty dotted segment", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "Owner..Email")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error nil interface deref before index", func(t *testing.T) {
		t.Parallel()

		type box struct {
			V any
		}

		_, err := GetField(box{V: nil}, "V[0]")
		if !errors.Is(err, ErrPathNotFound) {
			t.Fatalf("expected ErrPathNotFound, got %v", err)
		}
	})

	t.Run("multiple indices", func(t *testing.T) {
		t.Parallel()

		type matrix struct {
			Rows [][]int
		}

		m := matrix{Rows: [][]int{{1, 2}, {3, 4}}}

		v, err := GetField(m, "Rows[1][0]")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := v.(int)
		if !ok || got != 3 {
			t.Fatalf("expected 3, got %v", v)
		}
	})
}

func TestErrValidation(t *testing.T) {
	t.Parallel()

	t.Run("wraps causes", func(t *testing.T) {
		t.Parallel()

		inner := errors.New("inner")
		err := ErrValidation(inner)

		if !errors.Is(err, ErrValidationFailed) {
			t.Fatalf("expected ErrValidationFailed, got %v", err)
		}

		if !errors.Is(err, inner) {
			t.Fatalf("expected inner cause wrapped, got %v", err)
		}
	})

	t.Run("error string format", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation(ErrFieldRequired)
		msg := err.Error()
		if !strings.Contains(msg, "validation") {
			t.Fatalf("expected message to contain validation, got %q", msg)
		}
	})
}
