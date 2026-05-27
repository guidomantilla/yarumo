package cast

import (
	"time"

	castinternal "github.com/spf13/cast"
)

// ToBool casts any value to a bool.
func ToBool(i any) (bool, error) {
	return castinternal.ToBoolE(i)
}

// ToString casts any value to a string.
func ToString(i any) (string, error) {
	return castinternal.ToStringE(i)
}

// ToTime casts any value to a time.Time.
func ToTime(i any) (time.Time, error) {
	return castinternal.ToTimeE(i)
}

// ToTimeInDefaultLocation casts any value to a time.Time in the specified location.
func ToTimeInDefaultLocation(i any, location *time.Location) (time.Time, error) {
	return castinternal.ToTimeInDefaultLocationE(i, location)
}

// ToDuration casts any value to a time.Duration.
func ToDuration(i any) (time.Duration, error) {
	return castinternal.ToDurationE(i)
}

// ToInt casts any value to an int type.
func ToInt(i any) (int, error) {
	return castinternal.ToIntE(i)
}

// ToInt8 casts any value to an int8 type.
func ToInt8(i any) (int8, error) {
	return castinternal.ToInt8E(i)
}

// ToInt16 casts any value to an int16 type.
func ToInt16(i any) (int16, error) {
	return castinternal.ToInt16E(i)
}

// ToInt32 casts any value to an int32 type.
func ToInt32(i any) (int32, error) {
	return castinternal.ToInt32E(i)
}

// ToInt64 casts any value to an int64 type.
func ToInt64(i any) (int64, error) {
	return castinternal.ToInt64E(i)
}

// ToUint casts any value to a uint type.
func ToUint(i any) (uint, error) {
	return castinternal.ToUintE(i)
}

// ToUint8 casts any value to a uint8 type.
func ToUint8(i any) (uint8, error) {
	return castinternal.ToUint8E(i)
}

// ToUint16 casts any value to a uint16 type.
func ToUint16(i any) (uint16, error) {
	return castinternal.ToUint16E(i)
}

// ToUint32 casts any value to a uint32 type.
func ToUint32(i any) (uint32, error) {
	return castinternal.ToUint32E(i)
}

// ToUint64 casts any value to a uint64 type.
func ToUint64(i any) (uint64, error) {
	return castinternal.ToUint64E(i)
}

// ToFloat32 casts any value to a float32 type.
func ToFloat32(i any) (float32, error) {
	return castinternal.ToFloat32E(i)
}

// ToFloat64 casts any value to a float64 type.
func ToFloat64(i any) (float64, error) {
	return castinternal.ToFloat64E(i)
}

// ToStringMapString casts any value to a map[string]string type.
func ToStringMapString(i any) map[string]string {
	return castinternal.ToStringMapString(i)
}

// ToStringMapStringSlice casts any value to a map[string][]string type.
func ToStringMapStringSlice(i any) map[string][]string {
	return castinternal.ToStringMapStringSlice(i)
}

// ToStringMapBool casts any value to a map[string]bool type.
func ToStringMapBool(i any) map[string]bool {
	return castinternal.ToStringMapBool(i)
}

// ToStringMapInt casts any value to a map[string]int type.
func ToStringMapInt(i any) map[string]int {
	return castinternal.ToStringMapInt(i)
}

// ToStringMapInt64 casts any value to a map[string]int64 type.
func ToStringMapInt64(i any) map[string]int64 {
	return castinternal.ToStringMapInt64(i)
}

// ToStringMap casts any value to a map[string]any type.
func ToStringMap(i any) (map[string]any, error) {
	return castinternal.ToStringMapE(i)
}

// ToSlice casts any value to a []any type.
func ToSlice(i any) ([]any, error) {
	return castinternal.ToSliceE(i)
}

// ToBoolSlice casts any value to a []bool type.
func ToBoolSlice(i any) ([]bool, error) {
	return castinternal.ToBoolSliceE(i)
}

// ToStringSlice casts any value to a []string type.
func ToStringSlice(i any) ([]string, error) {
	return castinternal.ToStringSliceE(i)
}

// ToIntSlice casts any value to a []int type.
func ToIntSlice(i any) ([]int, error) {
	return castinternal.ToIntSliceE(i)
}

// ToInt8Slice casts any value to a []int8 type.
func ToInt8Slice(i any) ([]int8, error) {
	return castinternal.ToInt8SliceE(i)
}

// ToInt16Slice casts any value to a []int16 type.
func ToInt16Slice(i any) ([]int16, error) {
	return castinternal.ToInt16SliceE(i)
}

// ToInt32Slice casts any value to a []int32 type.
func ToInt32Slice(i any) ([]int32, error) {
	return castinternal.ToInt32SliceE(i)
}

// ToInt64Slice casts any value to a []int64 type.
func ToInt64Slice(i any) ([]int64, error) {
	return castinternal.ToInt64SliceE(i)
}

// ToUintSlice casts any value to a []uint type.
func ToUintSlice(i any) ([]uint, error) {
	return castinternal.ToUintSliceE(i)
}

// ToUint8Slice casts any value to a []uint8 type.
func ToUint8Slice(i any) ([]uint8, error) {
	return castinternal.ToUint8SliceE(i)
}

// ToUint16Slice casts any value to a []uint16 type.
func ToUint16Slice(i any) ([]uint16, error) {
	return castinternal.ToUint16SliceE(i)
}

// ToUint32Slice casts any value to a []uint32 type.
func ToUint32Slice(i any) ([]uint32, error) {
	return castinternal.ToUint32SliceE(i)
}

// ToUint64Slice casts any value to a []uint64 type.
func ToUint64Slice(i any) ([]uint64, error) {
	return castinternal.ToUint64SliceE(i)
}

// ToFloat32Slice casts any value to a []float32 type.
func ToFloat32Slice(i any) ([]float32, error) {
	return castinternal.ToFloat32SliceE(i)
}

// ToFloat64Slice casts any value to a []float64 type.
func ToFloat64Slice(i any) ([]float64, error) {
	return castinternal.ToFloat64SliceE(i)
}

// ToDurationSlice casts any value to a []time.Duration type.
func ToDurationSlice(i any) ([]time.Duration, error) {
	return castinternal.ToDurationSliceE(i)
}
