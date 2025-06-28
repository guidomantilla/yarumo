package utils

import (
	"reflect"

	"github.com/google/go-cmp/cmp"

	"github.com/guidomantilla/yarumo/pkg/pointer"
)

func Equal(x any, y any) bool {
	return cmp.Equal(x, y)
}

func NotEqual(x any, y any) bool {
	return !cmp.Equal(x, y)
}

func Nil(x any) bool {
	return pointer.IsNil(x)
}

func NotNil(x any) bool {
	return pointer.IsNotNil(x)
}

func Empty(x any) bool {
	if x == nil {
		return true
	}

	val := reflect.ValueOf(x)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return true
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return val.Len() == 0
	case reflect.Interface:
		return val.IsNil()
	default:
		return pointer.IsZero(val)
	}
}

func NotEmpty(x any) bool {
	return !Empty(x)
}
