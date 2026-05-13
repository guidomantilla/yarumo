package hmacs

import (
	"crypto/hmac"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	crandom "github.com/guidomantilla/yarumo/common/random"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Digest is the recommended entry point for callers that receive the
// algorithm name as a string (e.g. loaded from config, a request header, or
// a database column). It performs a single registry Get and forwards to
// Method.Digest, returning ErrAlgorithmNotSupported when name is not
// registered.
//
// For callers that already hold a *Method (predefined or returned by Get),
// use Method.Digest directly; Digest exists purely to collapse the
// "Get + Digest" boilerplate at the config↔runtime seam.
func Digest(name string, key, data ctypes.Bytes) (ctypes.Bytes, error) {
	method, err := Get(name)
	if err != nil {
		return nil, err
	}

	return method.Digest(key, data)
}

// Validate is the recommended entry point for callers that receive the
// algorithm name as a string. It performs a single registry Get and forwards
// to Method.Validate, returning ErrAlgorithmNotSupported when name is not
// registered.
func Validate(name string, key, digest, data ctypes.Bytes) (bool, error) {
	method, err := Get(name)
	if err != nil {
		return false, err
	}

	return method.Validate(key, digest, data)
}

func key(method *Method) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	out, err := crandom.Bytes(method.keySize)
	if err != nil {
		return nil, cerrs.Wrap(ErrKeyGenerationFailed, err)
	}

	return out, nil
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
