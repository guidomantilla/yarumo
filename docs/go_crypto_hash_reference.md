# Algoritmos de Cifrado y Funciones Hash Seguras en Go

## 🔐 Algoritmos seguros en Go (`crypto`)

| Algoritmo         | Modo recomendado | Longitud | Uso común                    |
|-------------------|------------------|----------|------------------------------|
| AES-256           | GCM              | 32 bytes | Cifrado general, TLS         |
| ChaCha20-Poly1305 | AEAD             | 32 bytes | IoT, móviles, WireGuard VPN  |

## 🔁 Funciones hash criptográficas generales

| Algoritmo     | Paquete                        | Longitud | Seguro hoy | Observaciones                      |
|---------------|--------------------------------|----------|------------|-------------------------------------|
| SHA-256       | `crypto/sha256`                | 32 bytes | ✅ Sí       | Estándar en TLS, HMAC, blockchain   |
| SHA-512       | `crypto/sha512`                | 64 bytes | ✅ Sí       | Alta seguridad, más pesado          |
| SHA3-256      | `golang.org/x/crypto/sha3`     | 32 bytes | ✅ Sí       | SHA-3 ganador (Keccak)              |
| SHA3-512      | `golang.org/x/crypto/sha3`     | 64 bytes | ✅ Sí       | Más lento, pero robusto             |
| BLAKE2b-512   | `golang.org/x/crypto/blake2b`  | 64 bytes | ✅ Sí       | Muy rápido, alta resistencia         |
