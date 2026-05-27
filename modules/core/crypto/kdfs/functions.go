package kdfs

import (
	"io"

	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
	ctypes "github.com/guidomantilla/yarumo/core/common/types"
)

// hkdfDerive performs RFC 5869 HKDF (extract-and-expand) using the method's
// hash kind. salt and info may both be empty (nil), matching RFC 5869.
func hkdfDerive(method *Method, secret, salt, info ctypes.Bytes, length int) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if secret == nil {
		return nil, ErrSecretIsNil
	}

	if length <= 0 {
		return nil, ErrLengthInvalid
	}

	if !method.kind.Available() {
		return nil, ErrHashNotAvailable
	}

	reader := hkdf.New(method.kind.New, secret, salt, info)

	out := make([]byte, length)

	_, err := io.ReadFull(reader, out)
	if err != nil {
		return nil, cerrs.Wrap(ErrDeriveFailed, err)
	}

	return out, nil
}

// pbkdf2Derive performs RFC 8018 PBKDF2 using the method's HMAC hash and the
// iteration count stored in pbkdf2Params. info is ignored.
func pbkdf2Derive(method *Method, secret, salt, _ ctypes.Bytes, length int) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if secret == nil {
		return nil, ErrSecretIsNil
	}

	if salt == nil {
		return nil, ErrSaltIsNil
	}

	if length <= 0 {
		return nil, ErrLengthInvalid
	}

	if method.pbkdf2Params == nil {
		return nil, ErrParamsMissing
	}

	if !method.kind.Available() {
		return nil, ErrHashNotAvailable
	}

	out := pbkdf2.Key(secret, salt, method.pbkdf2Params.iterations, length, method.kind.New)

	return out, nil
}

// scryptDerive performs RFC 7914 scrypt using the method's scryptParams.
// info is ignored.
func scryptDerive(method *Method, secret, salt, _ ctypes.Bytes, length int) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if secret == nil {
		return nil, ErrSecretIsNil
	}

	if salt == nil {
		return nil, ErrSaltIsNil
	}

	if length <= 0 {
		return nil, ErrLengthInvalid
	}

	if method.scryptParams == nil {
		return nil, ErrParamsMissing
	}

	out, err := scrypt.Key(secret, salt, method.scryptParams.n, method.scryptParams.r, method.scryptParams.p, length)
	if err != nil {
		return nil, cerrs.Wrap(ErrDeriveFailed, err)
	}

	return out, nil
}
