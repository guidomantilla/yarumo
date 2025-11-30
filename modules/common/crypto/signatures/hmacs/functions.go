package hmacs

import (
	"crypto/hmac"

	"github.com/guidomantilla/yarumo/common/random"
	"github.com/guidomantilla/yarumo/common/types"
)

func key(size int) types.Bytes {
	return random.Key(size)
}

func digest(method *Method, key types.Bytes, data types.Bytes) (types.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
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

func validate(method *Method, key types.Bytes, digest_ types.Bytes, data types.Bytes) (bool, error) {
	if method == nil {
		return false, ErrMethodIsNil
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
