certs: x.509
 - parseo desde PEM/DER
 - validacion de cadenas
 - helpers para tls
 - agnostico de BD, HSM, etc

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


# RSASSA_PSS_SHA256 — Equivalencias

| Nombre | Significado / Descripción |
|--------|---------------------------|
| **RSASSA_PSS_SHA256** | Nombre oficial usado en AWS KMS, PKCS#11 y APIs modernas |
| **PS256** | Nombre oficial en JOSE/JWT/JWS (RFC 7518) |
| **RSASSA-PSS using SHA-256** | Nombre formal según PKCS#1 v2.2 |
| **RSASSA-PSS + SHA256 + MGF1(SHA256)** | Definición criptográfica precisa |
| **RSA-PSS(SHA256)** | Nombre genérico en literatura criptográfica |
| **sha256WithRSAandMGF1** | Nombre en certificados X.509 y CSR |
| **OID: 1.2.840.113549.1.1.10** | Identificador ASN.1 para RSASSA-PSS |
| **OpenSSL (CLI)** | `rsa_padding_mode:pss` + `-sha256` |
| **WebCrypto API** | `{ name: "RSA-PSS", hash: "SHA-256", saltLength: 32 }` |
| **Go (crypto/rsa)** | `rsa.SignPSS(..., crypto.SHA256, ...)` |
| **Java JCA/JCE** | `"RSASSA-PSS"` con parámetros `SHA-256` |
| **Key size:** variable | 2048/3072/4096 bits según clave RSA |
| **Signature size:** igual al tamaño de la clave RSA | Ej: 256 bytes (RSA 2048), 384 bytes (RSA 3072) |


# RSASSA_PSS_SHA512 — Equivalencias

| Nombre | Significado / Descripción |
|--------|---------------------------|
| **RSASSA_PSS_SHA512** | Nombre oficial usado en AWS KMS, PKCS#11 y APIs modernas |
| **PS512** | Nombre oficial en JOSE/JWT/JWS (RFC 7518) |
| **RSASSA-PSS using SHA-512** | Nombre formal según PKCS#1 v2.2 |
| **RSASSA-PSS + SHA512 + MGF1(SHA512)** | Definición criptográfica precisa |
| **RSA-PSS(SHA512)** | Nombre genérico en literatura criptográfica |
| **sha512WithRSAandMGF1** | Nombre en certificados X.509 y CSR |
| **OID: 1.2.840.113549.1.1.10** | Identificador ASN.1 para RSASSA-PSS (con hash=SHA512) |
| **OpenSSL (CLI)** | `rsa_padding_mode:pss` + `-sha512` |
| **WebCrypto API** | `{ name: "RSA-PSS", hash: "SHA-512", saltLength: 64 }` |
| **Go (crypto/rsa)** | `rsa.SignPSS(..., crypto.SHA512, ...)` |
| **Java JCA/JCE** | `"RSASSA-PSS"` con parámetros `SHA-512` |
| **Key size:** 3072 o 4096 bits recomendado | 2048 bits es desbalanceado |
| **Signature size:** igual al tamaño de la clave RSA | Ej: 512 bytes (RSA 4096) |


# Ed25519 — Equivalencias

| Nombre | Significado / Descripción |
|--------|---------------------------|
| **Ed25519** | Nombre oficial según RFC 8032 (EdDSA sobre Curve25519) |
| **ED25519** | Nombre usado en muchas APIs (Go, libsodium, OpenSSL ≥ 1.1.1) |
| **EdDSA with SHA-512 over Curve25519** | Descripción formal |
| **EdDSA + SHA-512 + edwards25519** | Nombre genérico |
| **OKP (Octet Key Pair) — Ed25519** | Nombre en JOSE/JWT/JWK (RFC 8037) |
| **"EdDSA" (alg) con crv:"Ed25519"** | Nombre en JWS/JWT (algoritmo EdDSA) |
| **OpenSSL** | `-ed25519` (firma y claves) |
| **Go (crypto/ed25519)** | `ed25519.Sign()` / `ed25519.Verify()` |
| **libsodium / NaCl** | `crypto_sign_ed25519_*` |
| **Key size:** 32 bytes | Clave pública Ed25519 |
| **Secret key size:** 32 bytes | Más 32 bytes internos de expansión |
| **Signature size:** 64 bytes | Tamaño fijo (R || S) |



# AES-128-GCM — Equivalent Names Across Standards

AES-128-GCM is one of the most widely used AEAD (Authenticated Encryption with Associated Data)
ciphers. Depending on the ecosystem or standard, it may appear under different names:

## ✔ Equivalent Names

| Context / Standard     | Name / Identifier                    |
|------------------------|---------------------------------------|
| Generic cryptography   | **AES-128-GCM**                       |
| NIST                   | **AES-GCM (128-bit key)**            |
| TLS 1.2 / 1.3          | `AES_128_GCM` / `TLS_AES_128_GCM_SHA256` |
| JOSE / JWE             | `A128GCM`                             |
| HPKE (RFC 9180)        | `AES-128-GCM` (AEAD ID = 0x01)        |
| AWS Encryption SDK     | `AES/GCM/NoPadding (128-bit key)`     |
| libsodium              | *(not used — prefers ChaCha20)*       |
| Google Tink            | `AesGcmKey`, keySize = 16 bytes       |
| OpenSSL                | `aes-128-gcm`                         |
| Java JCE               | `"AES/GCM/NoPadding"` (16-byte key)   |
| .NET                   | `AesGcm` (KeySize = 128)              |
| Go (crypto/cipher)     | `cipher.NewGCM(block)` where block = `aes.NewCipher(key)` and `len(key)=16` |

## ✔ Key Size
- 16 bytes = 128 bits

## ✔ Nonce / IV
- 12 bytes (96 bits) recommended  
  (standard GCM nonce size)

## ✔ Tag (Authentication Tag)
- 16 bytes (128 bits)

---

# Summary
All names refer to the **same cipher**:
AES block cipher with 128-bit key, used in Galois/Counter Mode (GCM), providing authenticated encryption (AEAD).


# AES-256-GCM — Equivalent Names Across Standards

AES-256-GCM is a widely used AEAD (Authenticated Encryption with Associated Data) cipher
offering strong security and high performance (especially with AES-NI).

Depending on the standard or ecosystem, it appears under different names.

## ✔ Equivalent Names

| Context / Standard     | Name / Identifier                           |
|------------------------|----------------------------------------------|
| Generic cryptography   | **AES-256-GCM**                              |
| NIST                   | **AES-GCM (256-bit key)**                   |
| TLS 1.2 / 1.3          | `AES_256_GCM` / `TLS_AES_256_GCM_SHA384`     |
| JOSE / JWE             | `A256GCM`                                    |
| HPKE (RFC 9180)        | `AES-256-GCM` (AEAD ID = 0x03)               |
| AWS Encryption SDK     | `AES/GCM/NoPadding (256-bit key)`            |
| libsodium              | *(not used — prefers ChaCha20)*              |
| Google Tink            | `AesGcmKey`, keySize = 32 bytes              |
| OpenSSL                | `aes-256-gcm`                                |
| Java JCE               | `"AES/GCM/NoPadding"` (32-byte key)          |
| .NET                   | `AesGcm` (KeySize = 256)                     |
| Go (crypto/cipher)     | `cipher.NewGCM(block)` (block = aes.NewCipher(key) where len(key)=32) |

## ✔ Key Size
- 32 bytes = 256 bits

## ✔ Nonce / IV
- 12 bytes (96 bits) recommended (standard for GCM)

## ✔ Authentication Tag
- 16 bytes (128 bits)

---

# Summary

AES-256 in Galois/Counter Mode (GCM), providing:
- confidentiality,
- integrity,
- authenticity (AEAD).

It is considered a modern, secure, and widely interoperable symmetric encryption algorithm.


# ChaCha20-Poly1305 — Equivalent Generic Names

ChaCha20-Poly1305 is a modern AEAD (Authenticated Encryption with Associated Data)
cipher that combines the ChaCha20 stream cipher with the Poly1305 message authentication code.

Below are the canonical and equivalent generic names used across cryptography literature
and implementations.

## ✔ Generic / Equivalent Names

These names all refer to the exact same algorithm:

- **ChaCha20-Poly1305**
- **ChaCha20_Poly1305**
- **CHACHA20_POLY1305**
- **ChaCha20Poly1305**
- **ChaCha20 + Poly1305**
- **ChaCha20-Poly1305 AEAD**
- **AEAD_CHACHA20_POLY1305** (generic AEAD-style name)

All of these are semantically identical.

## ✔ Informal / Descriptive Names

Sometimes used in docs or libraries:

- **ChaCha20 with Poly1305 MAC**
- **ChaCha20 authenticated encryption**
- **ChaCha20-Poly1305 AE**
- **ChaCha20/Poly1305**

These are still referring to the same algorithm.

## ✔ Do NOT confuse with

These are *related but distinct* algorithms:

- **XChaCha20-Poly1305** (extended nonce: 24 bytes)
- **ChaCha20** (stream cipher only)
- **Poly1305** (MAC only)

## ✔ Parameters

- **Key size:** 32 bytes (256 bits)
- **Nonce size:** 12 bytes (96 bits)
- **Tag size:** 16 bytes (128 bits)
- **AEAD:** Yes (authenticated encryption)

## ✔ Summary

ChaCha20-Poly1305 is a fast, secure AEAD construction widely used in TLS 1.3, QUIC,
Noise Protocol, libsodium, and modern cryptographic systems.

It offers excellent performance on CPUs without AES acceleration and constant-time
operation suitable for side-channel–resistant implementations.



# XChaCha20-Poly1305 — Generic Equivalent Names

These names all refer to the same AEAD cipher:

- XChaCha20-Poly1305
- XChaCha20_Poly1305
- XCHACHA20_POLY1305
- XChaCha20Poly1305
- XChaCha20 + Poly1305
- AEAD_XChaCha20_Poly1305

## Parameters
- Key: 32 bytes (256 bits)
- Nonce: 24 bytes (192 bits)
- Tag: 16 bytes (128 bits)
- Type: AEAD (authenticated encryption)

## Notes
- Do not confuse with ChaCha20-Poly1305 (12-byte nonce).
- XChaCha20-Poly1305 is used in libsodium, WireGuard, Noise Protocol.












cryptos: cifrado asimetrico y simetrico
 - Asimetrico: 
   - RSA-OAEP-SHA256
   - RSA-OAEP-SHA512
   - ECDH_P256_HKDF_SHA256_AESGCM
   - ECDH_P521_HKDF_SHA512_AESGCM
   - ECDH_P256_HKDF_SHA256_CHACHA20POLY1305
   - ECDH_P521_HKDF_SHA512_CHACHA20POLY1305
   - X25519_HKDF_SHA256_AESGCM
   - X25519_HKDF_SHA256_CHACHA20POLY1305
   - X25519_HKDF_SHA512_AESGCM
   - X25519_HKDF_SHA512_CHACHA20POLY1305


Capa de KDF y passwords
•	HKDF sobre SHA-256 y/o SHA-512.
•	Para contraseñas: Argon2id o scrypt (aunque no estén en stdlib, vale la pena envolverlas si importas libs externas).