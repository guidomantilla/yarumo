// Package cast provides type-safe casting functions that wrap the spf13/cast library.
package cast

import "time"

var (
	_ ToBoolFn                  = ToBool
	_ ToStringFn                = ToString
	_ ToTimeFn                  = ToTime
	_ ToTimeInDefaultLocationFn = ToTimeInDefaultLocation
	_ ToDurationFn              = ToDuration
	_ ToIntFn                   = ToInt
	_ ToInt8Fn                  = ToInt8
	_ ToInt16Fn                 = ToInt16
	_ ToInt32Fn                 = ToInt32
	_ ToInt64Fn                 = ToInt64
	_ ToUintFn                  = ToUint
	_ ToUint8Fn                 = ToUint8
	_ ToUint16Fn                = ToUint16
	_ ToUint32Fn                = ToUint32
	_ ToUint64Fn                = ToUint64
	_ ToFloat32Fn               = ToFloat32
	_ ToFloat64Fn               = ToFloat64
	_ ToStringMapStringFn       = ToStringMapString
	_ ToStringMapStringSliceFn  = ToStringMapStringSlice
	_ ToStringMapBoolFn         = ToStringMapBool
	_ ToStringMapIntFn          = ToStringMapInt
	_ ToStringMapInt64Fn        = ToStringMapInt64
	_ ToStringMapFn             = ToStringMap
	_ ToSliceFn                 = ToSlice
	_ ToBoolSliceFn             = ToBoolSlice
	_ ToStringSliceFn           = ToStringSlice
	_ ToIntSliceFn              = ToIntSlice
	_ ToInt8SliceFn             = ToInt8Slice
	_ ToInt16SliceFn            = ToInt16Slice
	_ ToInt32SliceFn            = ToInt32Slice
	_ ToInt64SliceFn            = ToInt64Slice
	_ ToUintSliceFn             = ToUintSlice
	_ ToUint8SliceFn            = ToUint8Slice
	_ ToUint16SliceFn           = ToUint16Slice
	_ ToUint32SliceFn           = ToUint32Slice
	_ ToUint64SliceFn           = ToUint64Slice
	_ ToFloat32SliceFn          = ToFloat32Slice
	_ ToFloat64SliceFn          = ToFloat64Slice
	_ ToDurationSliceFn         = ToDurationSlice
)

// ToBoolFn is the function type for ToBool.
type ToBoolFn func(i any) (bool, error)

// ToStringFn is the function type for ToString.
type ToStringFn func(i any) (string, error)

// ToTimeFn is the function type for ToTime.
type ToTimeFn func(i any) (time.Time, error)

// ToTimeInDefaultLocationFn is the function type for ToTimeInDefaultLocation.
type ToTimeInDefaultLocationFn func(i any, location *time.Location) (time.Time, error)

// ToDurationFn is the function type for ToDuration.
type ToDurationFn func(i any) (time.Duration, error)

// ToIntFn is the function type for ToInt.
type ToIntFn func(i any) (int, error)

// ToInt8Fn is the function type for ToInt8.
type ToInt8Fn func(i any) (int8, error)

// ToInt16Fn is the function type for ToInt16.
type ToInt16Fn func(i any) (int16, error)

// ToInt32Fn is the function type for ToInt32.
type ToInt32Fn func(i any) (int32, error)

// ToInt64Fn is the function type for ToInt64.
type ToInt64Fn func(i any) (int64, error)

// ToUintFn is the function type for ToUint.
type ToUintFn func(i any) (uint, error)

// ToUint8Fn is the function type for ToUint8.
type ToUint8Fn func(i any) (uint8, error)

// ToUint16Fn is the function type for ToUint16.
type ToUint16Fn func(i any) (uint16, error)

// ToUint32Fn is the function type for ToUint32.
type ToUint32Fn func(i any) (uint32, error)

// ToUint64Fn is the function type for ToUint64.
type ToUint64Fn func(i any) (uint64, error)

// ToFloat32Fn is the function type for ToFloat32.
type ToFloat32Fn func(i any) (float32, error)

// ToFloat64Fn is the function type for ToFloat64.
type ToFloat64Fn func(i any) (float64, error)

// ToStringMapStringFn is the function type for ToStringMapString.
type ToStringMapStringFn func(i any) map[string]string

// ToStringMapStringSliceFn is the function type for ToStringMapStringSlice.
type ToStringMapStringSliceFn func(i any) map[string][]string

// ToStringMapBoolFn is the function type for ToStringMapBool.
type ToStringMapBoolFn func(i any) map[string]bool

// ToStringMapIntFn is the function type for ToStringMapInt.
type ToStringMapIntFn func(i any) map[string]int

// ToStringMapInt64Fn is the function type for ToStringMapInt64.
type ToStringMapInt64Fn func(i any) map[string]int64

// ToStringMapFn is the function type for ToStringMap.
type ToStringMapFn func(i any) (map[string]any, error)

// ToSliceFn is the function type for ToSlice.
type ToSliceFn func(i any) ([]any, error)

// ToBoolSliceFn is the function type for ToBoolSlice.
type ToBoolSliceFn func(i any) ([]bool, error)

// ToStringSliceFn is the function type for ToStringSlice.
type ToStringSliceFn func(i any) ([]string, error)

// ToIntSliceFn is the function type for ToIntSlice.
type ToIntSliceFn func(i any) ([]int, error)

// ToInt8SliceFn is the function type for ToInt8Slice.
type ToInt8SliceFn func(i any) ([]int8, error)

// ToInt16SliceFn is the function type for ToInt16Slice.
type ToInt16SliceFn func(i any) ([]int16, error)

// ToInt32SliceFn is the function type for ToInt32Slice.
type ToInt32SliceFn func(i any) ([]int32, error)

// ToInt64SliceFn is the function type for ToInt64Slice.
type ToInt64SliceFn func(i any) ([]int64, error)

// ToUintSliceFn is the function type for ToUintSlice.
type ToUintSliceFn func(i any) ([]uint, error)

// ToUint8SliceFn is the function type for ToUint8Slice.
type ToUint8SliceFn func(i any) ([]uint8, error)

// ToUint16SliceFn is the function type for ToUint16Slice.
type ToUint16SliceFn func(i any) ([]uint16, error)

// ToUint32SliceFn is the function type for ToUint32Slice.
type ToUint32SliceFn func(i any) ([]uint32, error)

// ToUint64SliceFn is the function type for ToUint64Slice.
type ToUint64SliceFn func(i any) ([]uint64, error)

// ToFloat32SliceFn is the function type for ToFloat32Slice.
type ToFloat32SliceFn func(i any) ([]float32, error)

// ToFloat64SliceFn is the function type for ToFloat64Slice.
type ToFloat64SliceFn func(i any) ([]float64, error)

// ToDurationSliceFn is the function type for ToDurationSlice.
type ToDurationSliceFn func(i any) ([]time.Duration, error)
