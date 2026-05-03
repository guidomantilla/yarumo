package hmacs

import (
	"crypto/hmac"

	crandom "github.com/guidomantilla/yarumo/common/random"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func key(method *Method) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	return crandom.Bytes(method.keySize), nil
}

func digest(method *Method, key ctypes.Bytes, data ctypes.Bytes) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if key == nil {
		return nil, ErrKeyIsNil
	}

	if !method.kind.Available() {
		return nil, ErrHashNotAvailable
	}

	h := hmac.New(method.kind.New, key)

	_, err := h.Write(data)
	if err != nil {
		return nil, err
	}

	out := h.Sum(nil)

	return out, nil
}

func validate(method *Method, key ctypes.Bytes, digest_ ctypes.Bytes, data ctypes.Bytes) (bool, error) {
	if method == nil {
		return false, ErrMethodIsNil
	}

	if key == nil {
		return false, ErrKeyIsNil
	}

	if !method.kind.Available() {
		return false, ErrHashNotAvailable
	}

	calculated, err := digest(method, key, data)
	if err != nil {
		return false, err
	}

	ok := hmac.Equal(digest_, calculated)

	return ok, nil
}
