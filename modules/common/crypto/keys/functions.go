package keys

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"

	"github.com/guidomantilla/yarumo/common/random"
	"github.com/guidomantilla/yarumo/common/types"
	"golang.org/x/crypto/hkdf"
)

func Key(size int) (types.Bytes, error) {
	return random.Bytes(size), nil
}

func RsaKey(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}

func Ed25519key() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

func EcdsaKey(curve elliptic.Curve) (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(curve, rand.Reader)
}

func xx(method *Method, key types.Bytes, salt types.Bytes, info types.Bytes) {

	deriver := hkdf.New(hash.New, key, salt, info)
}
