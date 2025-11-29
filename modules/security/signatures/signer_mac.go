package signatures

import (
	"errors"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/errs"

	"github.com/guidomantilla/yarumo/security/signatures/macs"
)

type MacSigner struct {
	alg macs.Algorithm
}

func NewMacSigner(alg macs.Algorithm) *MacSigner {
	return &MacSigner{alg: alg}
}

func (s *MacSigner) Sign(key any, data []byte) ([]byte, error) {
	assert.NotNil(s, "signer is nil")

	keyBytes, ok := key.([]byte)
	if !ok {
		return nil, errs.Wrap(ErrInvalidKeyType, errors.New("HMAC_SHA256 sign expects []byte"))
	}
	return s.alg.Fn(keyBytes, data)
}

func (s *MacSigner) Verify(key any, signature []byte, data []byte) (bool, error) {
	assert.NotNil(s, "signer is nil")

	keyBytes, ok := key.([]byte)
	if !ok {
		return false, errs.Wrap(ErrInvalidKeyType, errors.New("HMAC_SHA256 sign expects []byte"))
	}

	return macs.Verify(keyBytes, signature, data, s.alg.Fn)
}
