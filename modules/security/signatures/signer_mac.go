package signatures

import (
	"crypto/hmac"
	"errors"

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

	keyBytes, ok := key.([]byte)
	if !ok {
		return nil, errs.Wrap(ErrInvalidKeyType, errors.New("HMAC_SHA256 sign expects []byte"))
	}
	return s.macFn(keyBytes, data)
}

func (s *MacSigner) Verify(key any, signature []byte, data []byte) error {

	sig, err := s.Sign(key, data)
	if err != nil {
		return err
	}

	if !hmac.Equal(signature, sig) {
		return ErrSignatureInvalid
	}

	return nil
}
