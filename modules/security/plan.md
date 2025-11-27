certs: x.509
 - parseo desde PEM/DER
 - validacion de cadenas
 - helpers para tls
 - agnostico de BD, HSM, etc


signatures: Ed25519, ECDSA_P256_SHA256, RSASSA_PSS_SHA256

cryptos: cifrado asimetrico y simetrico
 - Asimetrico: 
   - RSA-OAEP-SHA256
   - ECDH_P256_HKDF_SHA256_AESGCM
   - X25519_HKDF_SHA256_AESGCM
   - X25519-HKDF-ChaCha20-Poly1305


keys: 
- para claves de 32 bytes AES-256-GCM y CHACHA20-POLY1305
- para claves de 64 bytes para KeyPairs

type SymmetricKey interface {
BaseKey
// Devuelve el material bruto.
// Quien use esto debe ser cuidadoso.
Bytes() []byte
}

type PublicKey interface {
BaseKey
// Representación pública (ej: para serializarla a PEM/DER en otro paquete)
Public() any
}

type PrivateKey interface {
BaseKey
// Devuelve el objeto de clave privada alg-específico (ed25519.PrivateKey, *ecdsa.PrivateKey, *rsa.PrivateKey, etc.)
Private() any
PublicKey() PublicKey
}