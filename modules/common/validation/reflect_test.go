package validation_test

import (
	"errors"
	"testing"

	cvalidation "github.com/guidomantilla/yarumo/common/validation"
)

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
