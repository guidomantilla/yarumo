package validation_test

import (
	"errors"
	"strings"
	"testing"

	cvalidation "github.com/guidomantilla/yarumo/common/validation"
)

func TestIsRequired(t *testing.T) {
	t.Parallel()

	t.Run("happy path string", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsRequired("hello")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path int", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsRequired(42)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty string", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsRequired("")
		if !errors.Is(err, cvalidation.ErrFieldRequired) {
			t.Fatalf("expected ErrFieldRequired, got %v", err)
		}
	})

	t.Run("error zero int", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsRequired(0)
		if !errors.Is(err, cvalidation.ErrFieldRequired) {
			t.Fatalf("expected ErrFieldRequired, got %v", err)
		}
	})
}

func TestMustBeUndefined(t *testing.T) {
	t.Parallel()

	t.Run("happy path empty string", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.MustBeUndefined("")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error non-empty string", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.MustBeUndefined("hello")
		if !errors.Is(err, cvalidation.ErrFieldMustBeUndefined) {
			t.Fatalf("expected ErrFieldMustBeUndefined, got %v", err)
		}
	})

	t.Run("happy path zero int", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.MustBeUndefined(0)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestMinLen(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.MinLen("hello", 3)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error below", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.MinLen("hi", 5)
		if !errors.Is(err, cvalidation.ErrMinLen) {
			t.Fatalf("expected ErrMinLen, got %v", err)
		}
	})

	t.Run("negative threshold accepts empty", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.MinLen("", -3)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestMaxLen(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.MaxLen("hello", 10)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error above", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.MaxLen("hello world", 5)
		if !errors.Is(err, cvalidation.ErrMaxLen) {
			t.Fatalf("expected ErrMaxLen, got %v", err)
		}
	})

	t.Run("negative threshold rejects non-empty", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.MaxLen("x", -1)
		if !errors.Is(err, cvalidation.ErrMaxLen) {
			t.Fatalf("expected ErrMaxLen, got %v", err)
		}
	})
}

func TestMatchesRegex(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.MatchesRegex("abc123", `^[a-z]+\d+$`)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error mismatch", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.MatchesRegex("ABC", `^[a-z]+$`)
		if !errors.Is(err, cvalidation.ErrRegexMismatch) {
			t.Fatalf("expected ErrRegexMismatch, got %v", err)
		}
	})

	t.Run("error invalid pattern", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.MatchesRegex("x", `[`)
		if !errors.Is(err, cvalidation.ErrRegexInvalid) {
			t.Fatalf("expected ErrRegexInvalid, got %v", err)
		}
	})
}

func TestIsEmail(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsEmail("a@b.com")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsEmail("")
		if !errors.Is(err, cvalidation.ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid, got %v", err)
		}
	})

	t.Run("error malformed", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsEmail("not-an-email")
		if !errors.Is(err, cvalidation.ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid, got %v", err)
		}
	})

	t.Run("error with display name", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsEmail("Name <a@b.com>")
		if !errors.Is(err, cvalidation.ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid, got %v", err)
		}
	})
}

func TestIsURL(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsURL("https://example.com/path")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsURL("")
		if !errors.Is(err, cvalidation.ErrURLInvalid) {
			t.Fatalf("expected ErrURLInvalid, got %v", err)
		}
	})

	t.Run("error missing scheme", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsURL("example.com")
		if !errors.Is(err, cvalidation.ErrURLInvalid) {
			t.Fatalf("expected ErrURLInvalid, got %v", err)
		}
	})

	t.Run("error unparseable", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsURL("http://[::1")
		if !errors.Is(err, cvalidation.ErrURLInvalid) {
			t.Fatalf("expected ErrURLInvalid, got %v", err)
		}
	})
}

func TestMin(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.Min(10, 5)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error below", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.Min(3, 10)
		if !errors.Is(err, cvalidation.ErrMinValue) {
			t.Fatalf("expected ErrMinValue, got %v", err)
		}
	})

	t.Run("happy path float", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.Min(1.5, 1.0)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestMax(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.Max(3, 10)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error above", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.Max(20, 10)
		if !errors.Is(err, cvalidation.ErrMaxValue) {
			t.Fatalf("expected ErrMaxValue, got %v", err)
		}
	})
}

func TestInRange(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.InRange(5, 0, 10)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error below", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.InRange(-1, 0, 10)
		if !errors.Is(err, cvalidation.ErrOutOfRange) {
			t.Fatalf("expected ErrOutOfRange, got %v", err)
		}
	})

	t.Run("error above", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.InRange(20, 0, 10)
		if !errors.Is(err, cvalidation.ErrOutOfRange) {
			t.Fatalf("expected ErrOutOfRange, got %v", err)
		}
	})

	t.Run("error invalid range", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.InRange(5, 10, 0)
		if !errors.Is(err, cvalidation.ErrInvalidRange) {
			t.Fatalf("expected ErrInvalidRange, got %v", err)
		}
	})
}

func TestIsUUID(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsUUID("550e8400-e29b-41d4-a716-446655440000")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsUUID("")
		if !errors.Is(err, cvalidation.ErrUUIDInvalid) {
			t.Fatalf("expected ErrUUIDInvalid, got %v", err)
		}
	})

	t.Run("error malformed", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsUUID("not-a-uuid")
		if !errors.Is(err, cvalidation.ErrUUIDInvalid) {
			t.Fatalf("expected ErrUUIDInvalid, got %v", err)
		}
	})
}

func TestIsULID(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsULID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsULID("")
		if !errors.Is(err, cvalidation.ErrULIDInvalid) {
			t.Fatalf("expected ErrULIDInvalid, got %v", err)
		}
	})

	t.Run("error malformed", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.IsULID("not-a-ulid")
		if !errors.Is(err, cvalidation.ErrULIDInvalid) {
			t.Fatalf("expected ErrULIDInvalid, got %v", err)
		}
	})
}

func TestNonEmpty(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.NonEmpty([]int{1, 2, 3})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.NonEmpty([]int{})
		if !errors.Is(err, cvalidation.ErrCollectionEmpty) {
			t.Fatalf("expected ErrCollectionEmpty, got %v", err)
		}
	})

	t.Run("error nil slice", func(t *testing.T) {
		t.Parallel()

		var xs []string

		err := cvalidation.NonEmpty(xs)
		if !errors.Is(err, cvalidation.ErrCollectionEmpty) {
			t.Fatalf("expected ErrCollectionEmpty, got %v", err)
		}
	})
}

func TestEach(t *testing.T) {
	t.Parallel()

	t.Run("happy path all pass", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.Each([]string{"a@b.com", "c@d.com"}, cvalidation.IsEmail)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error one fails", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.Each([]string{"a@b.com", "not-an-email"}, cvalidation.IsEmail)
		if !errors.Is(err, cvalidation.ErrEachFailed) {
			t.Fatalf("expected ErrEachFailed, got %v", err)
		}

		if !errors.Is(err, cvalidation.ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid to be wrapped, got %v", err)
		}
	})

	t.Run("nil check no-ops", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.Each([]int{1, 2, 3}, nil)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("empty slice trivially passes", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.Each([]int{}, func(int) error { return errors.New("never called") })
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

		v, err := cvalidation.GetField(p, "Owner.Email")
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

		v, err := cvalidation.GetField(p, "IDs[1]")
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

		v, err := cvalidation.GetField(p, "Owner.Tags[0]")
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

		v, err := cvalidation.GetField(p, "Name")
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

		v, err := cvalidation.GetField(m, "key")
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

		_, err := cvalidation.GetField(nil, "X")
		if !errors.Is(err, cvalidation.ErrObjectNil) {
			t.Fatalf("expected ErrObjectNil, got %v", err)
		}
	})

	t.Run("error empty path", func(t *testing.T) {
		t.Parallel()

		_, err := cvalidation.GetField(pokemon{}, "")
		if !errors.Is(err, cvalidation.ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error missing field", func(t *testing.T) {
		t.Parallel()

		_, err := cvalidation.GetField(pokemon{}, "Nope")
		if !errors.Is(err, cvalidation.ErrPathNotFound) {
			t.Fatalf("expected ErrPathNotFound, got %v", err)
		}
	})

	t.Run("error type mismatch", func(t *testing.T) {
		t.Parallel()

		_, err := cvalidation.GetField(pokemon{Name: "x"}, "Name.Inner")
		if !errors.Is(err, cvalidation.ErrPathTypeMismatch) {
			t.Fatalf("expected ErrPathTypeMismatch, got %v", err)
		}
	})

	t.Run("error index out of range", func(t *testing.T) {
		t.Parallel()

		_, err := cvalidation.GetField(pokemon{IDs: []int{1}}, "IDs[5]")
		if !errors.Is(err, cvalidation.ErrIndexOutOfRange) {
			t.Fatalf("expected ErrIndexOutOfRange, got %v", err)
		}
	})

	t.Run("error index on non-slice", func(t *testing.T) {
		t.Parallel()

		_, err := cvalidation.GetField(pokemon{Name: "x"}, "Name[0]")
		if !errors.Is(err, cvalidation.ErrPathTypeMismatch) {
			t.Fatalf("expected ErrPathTypeMismatch, got %v", err)
		}
	})

	t.Run("error malformed bracket", func(t *testing.T) {
		t.Parallel()

		_, err := cvalidation.GetField(pokemon{}, "IDs[")
		if !errors.Is(err, cvalidation.ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error empty index", func(t *testing.T) {
		t.Parallel()

		_, err := cvalidation.GetField(pokemon{}, "IDs[]")
		if !errors.Is(err, cvalidation.ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error non-numeric index", func(t *testing.T) {
		t.Parallel()

		_, err := cvalidation.GetField(pokemon{}, "IDs[abc]")
		if !errors.Is(err, cvalidation.ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error negative index", func(t *testing.T) {
		t.Parallel()

		_, err := cvalidation.GetField(pokemon{}, "IDs[-1]")
		if !errors.Is(err, cvalidation.ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error leading bracket", func(t *testing.T) {
		t.Parallel()

		_, err := cvalidation.GetField(pokemon{}, "[0]")
		if !errors.Is(err, cvalidation.ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error nil pointer in path", func(t *testing.T) {
		t.Parallel()

		type wrapper struct {
			P *pokemon
		}

		_, err := cvalidation.GetField(wrapper{}, "P.Name")
		if !errors.Is(err, cvalidation.ErrPathNotFound) {
			t.Fatalf("expected ErrPathNotFound, got %v", err)
		}
	})

	t.Run("error map non-string key", func(t *testing.T) {
		t.Parallel()

		m := map[int]string{1: "x"}

		_, err := cvalidation.GetField(m, "1")
		if !errors.Is(err, cvalidation.ErrPathTypeMismatch) {
			t.Fatalf("expected ErrPathTypeMismatch, got %v", err)
		}
	})

	t.Run("error map missing key", func(t *testing.T) {
		t.Parallel()

		m := map[string]any{"a": 1}

		_, err := cvalidation.GetField(m, "missing")
		if !errors.Is(err, cvalidation.ErrPathNotFound) {
			t.Fatalf("expected ErrPathNotFound, got %v", err)
		}
	})

	t.Run("error trailing chars after bracket", func(t *testing.T) {
		t.Parallel()

		_, err := cvalidation.GetField(pokemon{}, "IDs[0]junk")
		if !errors.Is(err, cvalidation.ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error empty dotted segment", func(t *testing.T) {
		t.Parallel()

		_, err := cvalidation.GetField(pokemon{}, "Owner..Email")
		if !errors.Is(err, cvalidation.ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error nil interface deref before index", func(t *testing.T) {
		t.Parallel()

		type box struct {
			V any
		}

		_, err := cvalidation.GetField(box{V: nil}, "V[0]")
		if !errors.Is(err, cvalidation.ErrPathNotFound) {
			t.Fatalf("expected ErrPathNotFound, got %v", err)
		}
	})

	t.Run("multiple indices", func(t *testing.T) {
		t.Parallel()

		type matrix struct {
			Rows [][]int
		}

		m := matrix{Rows: [][]int{{1, 2}, {3, 4}}}

		v, err := cvalidation.GetField(m, "Rows[1][0]")
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
		err := cvalidation.ErrValidation(inner)

		if !errors.Is(err, cvalidation.ErrValidationFailed) {
			t.Fatalf("expected ErrValidationFailed, got %v", err)
		}

		if !errors.Is(err, inner) {
			t.Fatalf("expected inner cause wrapped, got %v", err)
		}
	})

	t.Run("error string format", func(t *testing.T) {
		t.Parallel()

		err := cvalidation.ErrValidation(cvalidation.ErrFieldRequired)
		msg := err.Error()
		if !strings.Contains(msg, "validation") {
			t.Fatalf("expected message to contain validation, got %q", msg)
		}
	})
}
