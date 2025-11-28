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


