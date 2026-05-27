// Package cast provides type-safe casting functions that wrap the spf13/cast
// library. The function-type aliases live in modules/common/cast/; this
// package only ships the concrete implementations and the compliance
// assertions tying each function to its alias.
package cast

import (
	ccast "github.com/guidomantilla/yarumo/core/common/cast"
)

var (
	_ ccast.ToBoolFn                  = ToBool
	_ ccast.ToStringFn                = ToString
	_ ccast.ToTimeFn                  = ToTime
	_ ccast.ToTimeInDefaultLocationFn = ToTimeInDefaultLocation
	_ ccast.ToDurationFn              = ToDuration
	_ ccast.ToIntFn                   = ToInt
	_ ccast.ToInt8Fn                  = ToInt8
	_ ccast.ToInt16Fn                 = ToInt16
	_ ccast.ToInt32Fn                 = ToInt32
	_ ccast.ToInt64Fn                 = ToInt64
	_ ccast.ToUintFn                  = ToUint
	_ ccast.ToUint8Fn                 = ToUint8
	_ ccast.ToUint16Fn                = ToUint16
	_ ccast.ToUint32Fn                = ToUint32
	_ ccast.ToUint64Fn                = ToUint64
	_ ccast.ToFloat32Fn               = ToFloat32
	_ ccast.ToFloat64Fn               = ToFloat64
	_ ccast.ToStringMapStringFn       = ToStringMapString
	_ ccast.ToStringMapStringSliceFn  = ToStringMapStringSlice
	_ ccast.ToStringMapBoolFn         = ToStringMapBool
	_ ccast.ToStringMapIntFn          = ToStringMapInt
	_ ccast.ToStringMapInt64Fn        = ToStringMapInt64
	_ ccast.ToStringMapFn             = ToStringMap
	_ ccast.ToSliceFn                 = ToSlice
	_ ccast.ToBoolSliceFn             = ToBoolSlice
	_ ccast.ToStringSliceFn           = ToStringSlice
	_ ccast.ToIntSliceFn              = ToIntSlice
	_ ccast.ToInt8SliceFn             = ToInt8Slice
	_ ccast.ToInt16SliceFn            = ToInt16Slice
	_ ccast.ToInt32SliceFn            = ToInt32Slice
	_ ccast.ToInt64SliceFn            = ToInt64Slice
	_ ccast.ToUintSliceFn             = ToUintSlice
	_ ccast.ToUint8SliceFn            = ToUint8Slice
	_ ccast.ToUint16SliceFn           = ToUint16Slice
	_ ccast.ToUint32SliceFn           = ToUint32Slice
	_ ccast.ToUint64SliceFn           = ToUint64Slice
	_ ccast.ToFloat32SliceFn          = ToFloat32Slice
	_ ccast.ToFloat64SliceFn          = ToFloat64Slice
	_ ccast.ToDurationSliceFn         = ToDurationSlice
)
