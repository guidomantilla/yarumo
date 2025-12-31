package cast

import (
	"reflect"
	"testing"
	"time"
)

func TestCastFunctions(t *testing.T) {
	// Scalars
	if v, err := ToBool(true); err != nil || v != true {
		t.Fatalf("ToBool failed: %v, %v", v, err)
	}

	if v, err := ToString("hello"); err != nil || v != "hello" {
		t.Fatalf("ToString failed: %v, %v", v, err)
	}

	now := time.Now().Truncate(time.Second) // truncate to avoid monotonic diffs
	if v, err := ToTime(now); err != nil || !v.Equal(now) {
		t.Fatalf("ToTime failed: %v, %v", v, err)
	}

	loc := time.FixedZone("MyZone", -5*60*60)
	if v, err := ToTimeInDefaultLocation("2024-05-10 12:34:56", loc); err != nil || v.Location() != loc {
		t.Fatalf("ToTimeInDefaultLocation failed: %v, %v", v, err)
	}

	if v, err := ToDuration("1h30m"); err != nil || v != time.Hour+30*time.Minute {
		t.Fatalf("ToDuration failed: %v, %v", v, err)
	}

	if v, err := ToInt("42"); err != nil || v != 42 {
		t.Fatalf("ToInt failed: %v, %v", v, err)
	}

	if v, err := ToInt8("-8"); err != nil || v != int8(-8) {
		t.Fatalf("ToInt8 failed: %v, %v", v, err)
	}

	if v, err := ToInt16("-16"); err != nil || v != int16(-16) {
		t.Fatalf("ToInt16 failed: %v, %v", v, err)
	}

	if v, err := ToInt32("-32"); err != nil || v != int32(-32) {
		t.Fatalf("ToInt32 failed: %v, %v", v, err)
	}

	if v, err := ToInt64("-64"); err != nil || v != int64(-64) {
		t.Fatalf("ToInt64 failed: %v, %v", v, err)
	}

	if v, err := ToUint("7"); err != nil || v != uint(7) {
		t.Fatalf("ToUint failed: %v, %v", v, err)
	}

	if v, err := ToUint8("8"); err != nil || v != uint8(8) {
		t.Fatalf("ToUint8 failed: %v, %v", v, err)
	}

	if v, err := ToUint16("16"); err != nil || v != uint16(16) {
		t.Fatalf("ToUint16 failed: %v, %v", v, err)
	}

	if v, err := ToUint32("32"); err != nil || v != uint32(32) {
		t.Fatalf("ToUint32 failed: %v, %v", v, err)
	}

	if v, err := ToUint64("64"); err != nil || v != uint64(64) {
		t.Fatalf("ToUint64 failed: %v, %v", v, err)
	}

	if v, err := ToFloat32("3.5"); err != nil || v != float32(3.5) {
		t.Fatalf("ToFloat32 failed: %v, %v", v, err)
	}

	if v, err := ToFloat64("6.25"); err != nil || v != float64(6.25) {
		t.Fatalf("ToFloat64 failed: %v, %v", v, err)
	}

	// Maps
	if v := ToStringMapString(map[string]any{"a": "1", "b": 2}); !reflect.DeepEqual(v, map[string]string{"a": "1", "b": "2"}) {
		t.Fatalf("ToStringMapString failed: %v", v)
	}

	if v := ToStringMapStringSlice(map[string]any{"a": []any{"x", "y"}}); !reflect.DeepEqual(v, map[string][]string{"a": []string{"x", "y"}}) {
		t.Fatalf("ToStringMapStringSlice failed: %v", v)
	}

	if v := ToStringMapBool(map[string]any{"t": true, "f": false, "s": "true"}); !reflect.DeepEqual(v, map[string]bool{"t": true, "f": false, "s": true}) {
		t.Fatalf("ToStringMapBool failed: %v", v)
	}

	if v := ToStringMapInt(map[string]any{"a": 1, "b": "2"}); !reflect.DeepEqual(v, map[string]int{"a": 1, "b": 2}) {
		t.Fatalf("ToStringMapInt failed: %v", v)
	}

	if v := ToStringMapInt64(map[string]any{"a": int64(1), "b": "2"}); !reflect.DeepEqual(v, map[string]int64{"a": 1, "b": 2}) {
		t.Fatalf("ToStringMapInt64 failed: %v", v)
	}

	if v, err := ToStringMap(map[string]any{"a": "b", "n": 1}); err != nil || !reflect.DeepEqual(v, map[string]any{"a": "b", "n": 1}) {
		t.Fatalf("ToStringMap failed: %v, %v", v, err)
	}

	// Slices (generic)
	if v, err := ToSlice([]any{1, "a", true}); err != nil || !reflect.DeepEqual(v, []any{1, "a", true}) {
		t.Fatalf("ToSlice failed: %v, %v", v, err)
	}

	// Slices (typed)
	if v, err := ToBoolSlice([]any{true, false, "true"}); err != nil || !reflect.DeepEqual(v, []bool{true, false, true}) {
		t.Fatalf("ToBoolSlice failed: %v, %v", v, err)
	}

	if v, err := ToStringSlice([]any{"a", 2}); err != nil || !reflect.DeepEqual(v, []string{"a", "2"}) {
		t.Fatalf("ToStringSlice failed: %v, %v", v, err)
	}

	if v, err := ToIntSlice([]any{1, "2"}); err != nil || !reflect.DeepEqual(v, []int{1, 2}) {
		t.Fatalf("ToIntSlice failed: %v, %v", v, err)
	}

	if v, err := ToInt8Slice([]any{1, "2"}); err != nil || !reflect.DeepEqual(v, []int8{1, 2}) {
		t.Fatalf("ToInt8Slice failed: %v, %v", v, err)
	}

	if v, err := ToInt16Slice([]any{1, "2"}); err != nil || !reflect.DeepEqual(v, []int16{1, 2}) {
		t.Fatalf("ToInt16Slice failed: %v, %v", v, err)
	}

	if v, err := ToInt32Slice([]any{1, "2"}); err != nil || !reflect.DeepEqual(v, []int32{1, 2}) {
		t.Fatalf("ToInt32Slice failed: %v, %v", v, err)
	}

	if v, err := ToInt64Slice([]any{1, "2"}); err != nil || !reflect.DeepEqual(v, []int64{1, 2}) {
		t.Fatalf("ToInt64Slice failed: %v, %v", v, err)
	}

	if v, err := ToUintSlice([]any{1, "2"}); err != nil || !reflect.DeepEqual(v, []uint{1, 2}) {
		t.Fatalf("ToUintSlice failed: %v, %v", v, err)
	}

	if v, err := ToUint8Slice([]any{1, "2"}); err != nil || !reflect.DeepEqual(v, []uint8{1, 2}) {
		t.Fatalf("ToUint8Slice failed: %v, %v", v, err)
	}

	if v, err := ToUint16Slice([]any{1, "2"}); err != nil || !reflect.DeepEqual(v, []uint16{1, 2}) {
		t.Fatalf("ToUint16Slice failed: %v, %v", v, err)
	}

	if v, err := ToUint32Slice([]any{1, "2"}); err != nil || !reflect.DeepEqual(v, []uint32{1, 2}) {
		t.Fatalf("ToUint32Slice failed: %v, %v", v, err)
	}

	if v, err := ToUint64Slice([]any{1, "2"}); err != nil || !reflect.DeepEqual(v, []uint64{1, 2}) {
		t.Fatalf("ToUint64Slice failed: %v, %v", v, err)
	}

	if v, err := ToFloat32Slice([]any{1.5, "2.5"}); err != nil || !reflect.DeepEqual(v, []float32{1.5, 2.5}) {
		t.Fatalf("ToFloat32Slice failed: %v, %v", v, err)
	}

	if v, err := ToFloat64Slice([]any{1.5, "2.5"}); err != nil || !reflect.DeepEqual(v, []float64{1.5, 2.5}) {
		t.Fatalf("ToFloat64Slice failed: %v, %v", v, err)
	}

	if v, err := ToDurationSlice([]any{"1s", "2s", 3 * time.Second}); err != nil || !reflect.DeepEqual(v, []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second}) {
		t.Fatalf("ToDurationSlice failed: %v, %v", v, err)
	}
}
