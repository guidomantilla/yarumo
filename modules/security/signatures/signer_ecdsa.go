package signatures

import (
	cecdsa "crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"math/big"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/errs"
	"github.com/guidomantilla/yarumo/security/hashes"
	"github.com/guidomantilla/yarumo/security/signatures/ecdsa"

	"github.com/guidomantilla/yarumo/security/signatures/macs"
)

type EcdsaSigner struct {
	alg ecdsa.Algorithm
}

func NewEcdsaSigner(alg ecdsa.Algorithm) *EcdsaSigner {
	return &EcdsaSigner{alg: alg}
}

func (s *EcdsaSigner) Sign(key any, data []byte) ([]byte, error) {
	assert.NotNil(s, "signer is nil")

	keyBytes, ok := key.(*cecdsa.PrivateKey)
	if !ok {
		return nil, errs.Wrap(ErrInvalidKeyType, errors.New("HMAC_SHA256 sign expects []byte"))
	}
	return s.alg.Fn(keyBytes, data)
}

func (s *EcdsaSigner) Verify(key any, signature []byte, data []byte) error {
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

func ECDSA_SHA256_Verify(pub *ecdsa.PublicKey, data, sig types.Bytes) bool {
	if pub == nil {
		return false
	}
	if pub.Curve != elliptic.P256() {
		return false
	}
	if len(sig) != 64 {
		return false
	}

	r := new(big.Int).SetBytes(sig[0:32])
	s := new(big.Int).SetBytes(sig[32:64])

	hash := hashes.SHA256(data)

	return ecdsa.Verify(pub, hash, r, s)
}
