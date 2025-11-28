package signatures

import (
	"errors"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/errs"
	"github.com/guidomantilla/yarumo/security/signatures/macs"
)

type MacSigner struct {
	macFn macs.MacFn
}

func NewMacSigner(macFn macs.MacFn) *MacSigner {
	return &MacSigner{macFn: macFn}
}

func (s *MacSigner) Sign(key any, data []byte) ([]byte, error) {
	assert.NotNil(s, "signer is nil")

	keyBytes, ok := key.([]byte)
	if !ok {
		return nil, errs.Wrap(ErrInvalidKeyType, errors.New("HMAC_SHA256 sign expects []byte"))
	}
	return s.macFn(keyBytes, data)
}

func (s *MacSigner) Verify(key any, signature []byte, data []byte) error {
	assert.NotNil(s, "signer is nil")

	sig, err := s.Sign(key, data)
	if err != nil {
		return err
	}

	if macs.NotEqual(signature, sig) {
		return ErrSignatureInvalid
	}

	return nil
}
