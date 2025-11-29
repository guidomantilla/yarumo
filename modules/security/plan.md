certs: x.509
 - parseo desde PEM/DER
 - validacion de cadenas
 - helpers para tls
 - agnostico de BD, HSM, etc


signatures: Ed25519, ECDSA_P256_SHA256, RSASSA_PSS_SHA256

# HS256 — Equivalencias

| Nombre | Significado / Descripción |
|--------|---------------------------|
| **HS256** | Nombre oficial en JOSE/JWT/JWS |
| **HMAC_SHA256** | HMAC usando SHA-256 |
| **HMAC with SHA-256** | Definición formal |
| **HMAC + SHA256** | Nombre genérico |
| **Key size recomendado:** ≥ 256 bits (32 bytes) | Tamaño de la clave recomendado |
| **Output size:** 256 bits (32 bytes) | Tamaño del HMAC |


# HS512 — Equivalencias

| Nombre | Significado / Descripción |
|--------|---------------------------|
| **HS512** | Nombre oficial en JOSE/JWT/JWS |
| **HMAC_SHA512** | HMAC usando SHA-512 |
| **HMAC with SHA-512** | Definición formal |
| **HMAC + SHA512** | Nombre genérico |
| **Key size recomendado:** ≥ 512 bits (64 bytes) | Tamaño de la clave recomendado |
| **Output size:** 512 bits (64 bytes) | Tamaño del HMAC |

# ES256 — Equivalencias

| Nombre | Significado / Descripción |
|--------|---------------------------|
| **ES256** | Nombre oficial en JOSE/JWT/JWS |
| **ECDSA_P256_SHA256** | ECDSA sobre la curva P-256 usando SHA-256 |
| **ECDSA with SHA-256 over P-256** | Definición formal |
| **ECDSA + SHA256 + P-256** | Nombre genérico |
| **ECDSA + SHA256 + secp256r1** | Nombre en OpenSSL |
| **ECDSA + SHA256 + prime256v1** | Nombre en TLS/X.509 |
| **Key size:** 256 bits (32 bytes) | Tamaño de la clave |
| **Signature size:** 64 bytes (`R=32` + `S=32`) | Formato R \|\| S |


# ES512 — Equivalencias

| Nombre | Significado / Descripción |
|--------|---------------------------|
| **ES512** | Nombre oficial en JOSE/JWT/JWS |
| **ECDSA_P521_SHA512** | ECDSA sobre la curva P-521 usando SHA-512 |
| **ECDSA with SHA-512 over P-521** | Definición formal |
| **ECDSA + SHA512 + P-521** | Nombre genérico |
| **ECDSA + SHA512 + secp521r1** | Nombre en OpenSSL |
| **ECDSA + SHA512 + prime521v1** | Nombre en TLS/X.509 |
| **Key size:** 521 bits (66 bytes) | Tamaño de la clave |
| **Signature size:** 132 bytes (`R=66` + `S=66`) | Formato R \|\| S |




cryptos: cifrado asimetrico y simetrico
 - Asimetrico: 
   - RSA-OAEP-SHA256
   - ECDH_P256_HKDF_SHA256_AESGCM
   - X25519_HKDF_SHA256_AESGCM
   - X25519-HKDF-ChaCha20-Poly1305


