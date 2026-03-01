package cast

import (
	"reflect"
	"testing"
	"time"
)

// --- Scalar functions ---

func TestToBool(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToBool(true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != true {
			t.Fatalf("got %v, want true", v)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToBool("not-a-bool")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToBool(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != false {
			t.Fatalf("got %v, want false", v)
		}
	})
}

func TestToString(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToString("hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != "hello" {
			t.Fatalf("got %q, want %q", v, "hello")
		}
	})

	t.Run("nil input returns empty string", func(t *testing.T) {
		t.Parallel()

		v, err := ToString(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != "" {
			t.Fatalf("got %q, want %q", v, "")
		}
	})
}

func TestToTime(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		now := time.Now().Truncate(time.Second)

		v, err := ToTime(now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !v.Equal(now) {
			t.Fatalf("got %v, want %v", v, now)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToTime("invalid")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToTime(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		zero := time.Time{}
		if !v.Equal(zero) {
			t.Fatalf("got %v, want zero time", v)
		}
	})
}

func TestToTimeInDefaultLocation(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		loc := time.FixedZone("Test", -5*60*60)

		v, err := ToTimeInDefaultLocation("2024-05-10 12:34:56", loc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v.Location() != loc {
			t.Fatalf("got location %v, want %v", v.Location(), loc)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		loc := time.FixedZone("Test", 0)

		_, err := ToTimeInDefaultLocation("invalid", loc)
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		loc := time.FixedZone("Test", 0)

		v, err := ToTimeInDefaultLocation(nil, loc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		zero := time.Time{}
		if !v.Equal(zero) {
			t.Fatalf("got %v, want zero time", v)
		}
	})
}

func TestToDuration(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToDuration("1h30m")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := time.Hour + 30*time.Minute
		if v != want {
			t.Fatalf("got %v, want %v", v, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToDuration("not-a-duration")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToDuration(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != 0 {
			t.Fatalf("got %v, want 0", v)
		}
	})
}

// --- Integer functions ---

func TestToInt(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToInt("42")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != 42 {
			t.Fatalf("got %d, want 42", v)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt("not-a-number")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToInt(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != 0 {
			t.Fatalf("got %d, want 0", v)
		}
	})
}

func TestToInt8(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToInt8("-8")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != int8(-8) {
			t.Fatalf("got %d, want -8", v)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt8("not-a-number")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToInt8(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != int8(0) {
			t.Fatalf("got %d, want 0", v)
		}
	})
}

func TestToInt16(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToInt16("-16")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != int16(-16) {
			t.Fatalf("got %d, want -16", v)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt16("not-a-number")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToInt16(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != int16(0) {
			t.Fatalf("got %d, want 0", v)
		}
	})
}

func TestToInt32(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToInt32("-32")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != int32(-32) {
			t.Fatalf("got %d, want -32", v)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt32("not-a-number")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToInt32(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != int32(0) {
			t.Fatalf("got %d, want 0", v)
		}
	})
}

func TestToInt64(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToInt64("-64")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != int64(-64) {
			t.Fatalf("got %d, want -64", v)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt64("not-a-number")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToInt64(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != int64(0) {
			t.Fatalf("got %d, want 0", v)
		}
	})
}

// --- Unsigned integer functions ---

func TestToUint(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToUint("7")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != uint(7) {
			t.Fatalf("got %d, want 7", v)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUint("not-a-number")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToUint(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != uint(0) {
			t.Fatalf("got %d, want 0", v)
		}
	})
}

func TestToUint8(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToUint8("8")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != uint8(8) {
			t.Fatalf("got %d, want 8", v)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUint8("not-a-number")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToUint8(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != uint8(0) {
			t.Fatalf("got %d, want 0", v)
		}
	})
}

func TestToUint16(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToUint16("16")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != uint16(16) {
			t.Fatalf("got %d, want 16", v)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUint16("not-a-number")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToUint16(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != uint16(0) {
			t.Fatalf("got %d, want 0", v)
		}
	})
}

func TestToUint32(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToUint32("32")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != uint32(32) {
			t.Fatalf("got %d, want 32", v)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUint32("not-a-number")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToUint32(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != uint32(0) {
			t.Fatalf("got %d, want 0", v)
		}
	})
}

func TestToUint64(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToUint64("64")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != uint64(64) {
			t.Fatalf("got %d, want 64", v)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUint64("not-a-number")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToUint64(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != uint64(0) {
			t.Fatalf("got %d, want 0", v)
		}
	})
}

// --- Float functions ---

func TestToFloat32(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToFloat32("3.5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != float32(3.5) {
			t.Fatalf("got %v, want 3.5", v)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToFloat32("not-a-number")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToFloat32(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != float32(0) {
			t.Fatalf("got %v, want 0", v)
		}
	})
}

func TestToFloat64(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		v, err := ToFloat64("6.25")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != float64(6.25) {
			t.Fatalf("got %v, want 6.25", v)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToFloat64("not-a-number")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		t.Parallel()

		v, err := ToFloat64(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v != float64(0) {
			t.Fatalf("got %v, want 0", v)
		}
	})
}

// --- Map functions ---

func TestToStringMapString(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got := ToStringMapString(map[string]any{"a": "1", "b": 2})

		want := map[string]string{"a": "1", "b": "2"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("nil input returns empty map", func(t *testing.T) {
		t.Parallel()

		got := ToStringMapString(nil)

		if got == nil {
			t.Fatalf("got nil, want empty map")
		}

		if len(got) != 0 {
			t.Fatalf("got %v, want empty map", got)
		}
	})
}

func TestToStringMapStringSlice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got := ToStringMapStringSlice(map[string]any{"a": []any{"x", "y"}})

		want := map[string][]string{"a": {"x", "y"}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("nil input returns empty map", func(t *testing.T) {
		t.Parallel()

		got := ToStringMapStringSlice(nil)

		if got == nil {
			t.Fatalf("got nil, want empty map")
		}

		if len(got) != 0 {
			t.Fatalf("got %v, want empty map", got)
		}
	})
}

func TestToStringMapBool(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got := ToStringMapBool(map[string]any{"t": true, "f": false})

		want := map[string]bool{"t": true, "f": false}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("nil input returns empty map", func(t *testing.T) {
		t.Parallel()

		got := ToStringMapBool(nil)

		if got == nil {
			t.Fatalf("got nil, want empty map")
		}

		if len(got) != 0 {
			t.Fatalf("got %v, want empty map", got)
		}
	})
}

func TestToStringMapInt(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got := ToStringMapInt(map[string]any{"a": 1, "b": "2"})

		want := map[string]int{"a": 1, "b": 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("nil input returns nil map", func(t *testing.T) {
		t.Parallel()

		got := ToStringMapInt(nil)

		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})
}

func TestToStringMapInt64(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got := ToStringMapInt64(map[string]any{"a": int64(1), "b": "2"})

		want := map[string]int64{"a": 1, "b": 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("nil input returns nil map", func(t *testing.T) {
		t.Parallel()

		got := ToStringMapInt64(nil)

		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})
}

func TestToStringMap(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToStringMap(map[string]any{"a": "b", "n": 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := map[string]any{"a": "b", "n": 1}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToStringMap("not-a-map")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToStringMap(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

// --- Slice functions ---

func TestToSlice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToSlice([]any{1, "a", true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []any{1, "a", true}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToSlice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToSlice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToBoolSlice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToBoolSlice([]any{true, false, "true"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []bool{true, false, true}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToBoolSlice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToBoolSlice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToStringSlice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToStringSlice([]any{"a", 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []string{"a", "2"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToStringSlice(struct{}{})
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToStringSlice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToIntSlice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToIntSlice([]any{1, "2"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []int{1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToIntSlice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToIntSlice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToInt8Slice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToInt8Slice([]any{1, "2"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []int8{1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt8Slice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt8Slice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToInt16Slice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToInt16Slice([]any{1, "2"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []int16{1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt16Slice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt16Slice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToInt32Slice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToInt32Slice([]any{1, "2"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []int32{1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt32Slice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt32Slice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToInt64Slice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToInt64Slice([]any{1, "2"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []int64{1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt64Slice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt64Slice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToUintSlice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToUintSlice([]any{1, "2"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []uint{1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUintSlice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUintSlice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToUint8Slice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToUint8Slice([]any{1, "2"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []uint8{1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUint8Slice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUint8Slice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToUint16Slice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToUint16Slice([]any{1, "2"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []uint16{1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUint16Slice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUint16Slice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToUint32Slice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToUint32Slice([]any{1, "2"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []uint32{1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUint32Slice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUint32Slice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToUint64Slice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToUint64Slice([]any{1, "2"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []uint64{1, 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUint64Slice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToUint64Slice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToFloat32Slice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToFloat32Slice([]any{1.5, "2.5"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []float32{1.5, 2.5}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToFloat32Slice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToFloat32Slice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToFloat64Slice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToFloat64Slice([]any{1.5, "2.5"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []float64{1.5, 2.5}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToFloat64Slice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToFloat64Slice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}

func TestToDurationSlice(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		got, err := ToDurationSlice([]any{"1s", "2s", 3 * time.Second})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []time.Duration{time.Second, 2 * time.Second, 3 * time.Second}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToDurationSlice("not-a-slice")
		if err == nil {
			t.Fatalf("expected error for invalid input, got nil")
		}
	})

	t.Run("nil input returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ToDurationSlice(nil)
		if err == nil {
			t.Fatalf("expected error for nil input, got nil")
		}
	})
}
