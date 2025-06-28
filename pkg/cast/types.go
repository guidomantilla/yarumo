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

type ToBoolFn func(i any) (bool, error)

type ToStringFn func(i any) (string, error)

type ToTimeFn func(i any) (time.Time, error)

type ToTimeInDefaultLocationFn func(i any, location *time.Location) (time.Time, error)

type ToDurationFn func(i any) (time.Duration, error)

type ToIntFn func(i any) (int, error)

type ToInt8Fn func(i any) (int8, error)

type ToInt16Fn func(i any) (int16, error)

type ToInt32Fn func(i any) (int32, error)

type ToInt64Fn func(i any) (int64, error)

type ToUintFn func(i any) (uint, error)

type ToUint8Fn func(i any) (uint8, error)

type ToUint16Fn func(i any) (uint16, error)

type ToUint32Fn func(i any) (uint32, error)

type ToUint64Fn func(i any) (uint64, error)

type ToFloat32Fn func(i any) (float32, error)

type ToFloat64Fn func(i any) (float64, error)

type ToStringMapStringFn func(i any) map[string]string

type ToStringMapStringSliceFn func(i any) map[string][]string

type ToStringMapBoolFn func(i any) map[string]bool

type ToStringMapIntFn func(i any) map[string]int

type ToStringMapInt64Fn func(i any) map[string]int64

type ToStringMapFn func(i any) (map[string]any, error)

type ToSliceFn func(i any) ([]any, error)

type ToBoolSliceFn func(i any) ([]bool, error)

type ToStringSliceFn func(i any) ([]string, error)

type ToIntSliceFn func(i any) ([]int, error)

type ToInt8SliceFn func(i any) ([]int8, error)

type ToInt16SliceFn func(i any) ([]int16, error)

type ToInt32SliceFn func(i any) ([]int32, error)

type ToInt64SliceFn func(i any) ([]int64, error)

type ToUintSliceFn func(i any) ([]uint, error)

type ToUint8SliceFn func(i any) ([]uint8, error)

type ToUint16SliceFn func(i any) ([]uint16, error)

type ToUint32SliceFn func(i any) ([]uint32, error)

type ToUint64SliceFn func(i any) ([]uint64, error)

type ToFloat32SliceFn func(i any) ([]float32, error)

type ToFloat64SliceFn func(i any) ([]float64, error)

type ToDurationSliceFn func(i any) ([]time.Duration, error)
