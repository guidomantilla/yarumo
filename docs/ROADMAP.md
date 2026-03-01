# Roadmap

Trabajo planificado con diseño definido. Cada item tiene un plan de implementacion
detallado que cumple CODING_STANDARDS.md.

---

## crypto

### Opaque Tokens — `common/crypto/tokens`

Agregar soporte de opaque tokens al paquete `tokens` existente. Un opaque token es un token
encriptado donde el cliente no puede inspeccionar los claims — solo el servidor (que posee
la key) puede desencriptar y validar.

**Motivacion:**
- JWT es transparente: cualquiera puede decodificar los claims (solo la firma protege integridad, no confidencialidad).
- Opaque tokens resuelven: confidencialidad de claims, eliminacion de ataques de algorithm confusion, session tokens con metadata embebida invisible al cliente.
- Patron establecido: Fernet (Python), PASETO, Branca.
- Ya existe `ciphers/aead` con AES-GCM y XChaCha20-Poly1305. El opaque token lo reutiliza internamente.

**Diseno:**
El opaque token reutiliza la misma `Method` struct, los mismos `GenerateFn`/`ValidateFn` function types, y el mismo registry. No se crea un struct nuevo ni un paquete aparte.

**Flujo Generate:**
```
subject + payload + issuer + timeout
    -> construir claims map con iat/nbf/exp/iss/sub/payload
    -> json.Marshal(claims)
    -> aead.Method.Encrypt(key, jsonBytes, nil)
    -> base64url.RawEncoding.EncodeToString(ciphertext)
    -> return token string
```

**Flujo Validate:**
```
token string
    -> base64url.RawEncoding.DecodeString(token)
    -> aead.Method.Decrypt(key, ciphertext, nil)
    -> json.Unmarshal(plaintext, &claims)
    -> validar exp >= now, nbf <= now, iss == method.issuer (si configurado)
    -> return claims.Payload
```

**Formato del token:**
```
base64url( nonce || AEAD.Seal(key, nonce, json(claims), nil) )
```

Donde `claims` es:
```json
{
  "iss": "my-service",
  "sub": "user-123",
  "iat": 1709136000,
  "nbf": 1709136000,
  "exp": 1709222400,
  "payload": { "role": "admin" }
}
```

El nonce es generado internamente por `aead.Method.Encrypt` (ya lo prepend al ciphertext),
asi que el token opaco es simplemente el output de Encrypt codificado en base64url.

#### Plan de implementacion

##### Paso 1 — Nuevos fields en `Method` struct (`tokens.go`)

Agregar campo opcional para soportar opaque:

```go
type Method struct {
    name          string
    signingMethod jwt.SigningMethod  // nil para opaque
    signingKey    []byte
    verifyingKey  []byte
    issuer        string
    timeout       time.Duration
    generateFn    GenerateFn
    validateFn    ValidateFn
    cipher        *caead.Method     // nuevo: nil para JWT, non-nil para opaque
}
```

- `cipher` — puntero a un `aead.Method`. Si es nil, el method es JWT. Si no es nil, es opaque.
- Import alias: `caead "github.com/guidomantilla/yarumo/common/crypto/ciphers/aead"`.

##### Paso 2 — Nuevas funciones en `functions.go`

Agregar dos funciones privadas que implementan `GenerateFn` y `ValidateFn` para opaque:

```go
func generateOpaque(method *Method, subject string, payload Payload) (string, error)
func validateOpaque(method *Method, tokenString string) (Payload, error)
```

**`generateOpaque`:**
1. Validar `subject` no vacio (`cutils.Empty`), `payload` no nil (`cutils.Nil`).
2. Validar `method.signingKey` no nil y `method.cipher` no nil.
3. Construir un `opaqueClaims` map:
   - `"iss"`: `method.issuer` (solo si no vacio)
   - `"sub"`: `subject`
   - `"iat"`: `now.Unix()`
   - `"nbf"`: `now.Unix()`
   - `"exp"`: `now.Add(method.timeout).Unix()`
   - `"payload"`: `payload`
4. `json.Marshal(opaqueClaims)`.
5. `method.cipher.Encrypt(method.signingKey, jsonBytes, nil)`.
6. `base64.RawURLEncoding.EncodeToString(ciphered)`.
7. Return token string.

**`validateOpaque`:**
1. Validar `tokenString` no vacio.
2. `base64.RawURLEncoding.DecodeString(tokenString)`.
3. `method.cipher.Decrypt(method.signingKey, ciphered, nil)`.
4. `json.Unmarshal(plaintext, &opaqueClaims)` donde `opaqueClaims` es `map[string]any`.
5. Extraer y validar `"exp"` — debe ser `>= time.Now().Unix()`. Si expirado, retornar `ErrTokenExpired`.
6. Extraer y validar `"nbf"` — debe ser `<= time.Now().Unix()`. Si aun no valido, retornar `ErrTokenNotYetValid`.
7. Extraer y validar `"iss"` — si `method.issuer` no esta vacio, debe coincidir. Si no coincide, retornar `ErrTokenIssuerMismatch`.
8. Extraer `"payload"` como `Payload`. Si nil, retornar `ErrTokenPayloadEmpty`.
9. Return payload.

**Type compliance** — agregar en `types.go`:
```go
var (
    _ GenerateFn = generate
    _ ValidateFn = validate
    _ GenerateFn = generateOpaque
    _ ValidateFn = validateOpaque
)
```

##### Paso 3 — Nuevos sentinel errors en `errors.go`

```go
var (
    // ... existentes ...
    ErrTokenExpired         = errors.New("token expired")
    ErrTokenNotYetValid     = errors.New("token not yet valid")
    ErrTokenIssuerMismatch  = errors.New("token issuer mismatch")
    ErrCipherNil            = errors.New("cipher is nil")
    ErrTokenDecodeFailed    = errors.New("token decode failed")
    ErrTokenDecryptFailed   = errors.New("token decrypt failed")
    ErrTokenMarshalFailed   = errors.New("token marshal failed")
    ErrTokenUnmarshalFailed = errors.New("token unmarshal failed")
)
```

##### Paso 4 — Nuevas options en `options.go`

```go
// WithCipher sets the AEAD cipher method for opaque token generation.
func WithCipher(cipher *caead.Method) Option {
    return func(opts *Options) {
        if cipher != nil {
            opts.cipher = cipher
        }
    }
}
```

Agregar `cipher *caead.Method` al struct `Options`.
Actualizar `NewOptions` para NO setear un default de cipher (nil = JWT behavior, que es el default actual).

##### Paso 5 — Constructor `NewOpaqueMethod` en `tokens.go`

```go
// NewOpaqueMethod creates a new opaque token method backed by AEAD encryption.
func NewOpaqueMethod(name string, cipher *caead.Method, options ...Option) *Method {
    cassert.NotEmpty(name, "name is empty")
    cassert.NotNil(cipher, "cipher is nil")

    opts := NewOptions(
        append([]Option{
            WithCipher(cipher),
            WithGenerateFn(generateOpaque),
            WithValidateFn(validateOpaque),
        }, options...)...,
    )

    return &Method{
        name:       name,
        signingKey: opts.signingKey,
        issuer:     opts.issuer,
        timeout:    opts.timeout,
        generateFn: opts.generateFn,
        validateFn: opts.validateFn,
        cipher:     opts.cipher,
    }
}
```

- La key de `NewOptions` (random 64 bytes) se usa como encryption key. El usuario puede override con `WithKey(key)` pasando una key del tamano correcto para el cipher (16 o 32 bytes).
- `signingMethod` y `verifyingKey` quedan como zero values (no se usan en opaque).

##### Paso 6 — Predefined opaque method vars en `tokens.go`

```go
var (
    // JWT methods
    JWT_HS256 = NewMethod("JWT_HS256", jwt.SigningMethodHS256)
    JWT_HS384 = NewMethod("JWT_HS384", jwt.SigningMethodHS384)
    JWT_HS512 = NewMethod("JWT_HS512", jwt.SigningMethodHS512)

    // Opaque methods
    OPAQUE_AES_256_GCM        = NewOpaqueMethod("OPAQUE_AES_256_GCM", caead.AES_256_GCM, WithKey(crandom.Bytes(32)))
    OPAQUE_XCHACHA20_POLY1305 = NewOpaqueMethod("OPAQUE_XCHACHA20_POLY1305", caead.XCHACHA20_POLY1305, WithKey(crandom.Bytes(32)))
)
```

##### Paso 7 — Actualizar registry en `extensions.go`

```go
var methods = map[string]Method{
    JWT_HS256.name: *JWT_HS256,
    JWT_HS384.name: *JWT_HS384,
    JWT_HS512.name: *JWT_HS512,
    OPAQUE_AES_256_GCM.name:        *OPAQUE_AES_256_GCM,
    OPAQUE_XCHACHA20_POLY1305.name: *OPAQUE_XCHACHA20_POLY1305,
}
```

##### Paso 8 — Package doc en `types.go`

```go
// Package tokens provides JWT and opaque token generation and validation.
// JWT tokens use HMAC signing methods. Opaque tokens use AEAD encryption
// from the ciphers/aead package, producing tokens that clients cannot inspect.
package tokens
```

##### Paso 9 — Tests

**Nuevos tests para `generateOpaque` (en `functions_test.go`):**
- `TestGenerateOpaque` con subtests:
  - `t.Run("success")` — generate con AES-256-GCM, verificar que retorna string no vacio.
  - `t.Run("empty subject")` — subject vacio retorna `ErrSubjectEmpty`.
  - `t.Run("nil payload")` — payload nil retorna `ErrPayloadNil`.
  - `t.Run("nil signing key")` — signing key nil retorna `ErrSigningKeyNil`.
  - `t.Run("nil cipher")` — cipher nil retorna `ErrCipherNil`.

**Nuevos tests para `validateOpaque` (en `functions_test.go`):**
- `TestValidateOpaque` con subtests:
  - `t.Run("success")` — generate + validate roundtrip, verificar payload intacto.
  - `t.Run("empty token")` — token vacio retorna `ErrTokenEmpty`.
  - `t.Run("invalid base64")` — base64 invalido retorna error de decode.
  - `t.Run("tampered ciphertext")` — modificar un byte del ciphertext, verificar error de decrypt.
  - `t.Run("expired token")` — generar con timeout negativo o muy corto, esperar, validar.
  - `t.Run("issuer mismatch")` — generar con issuer A, validar con method con issuer B.
  - `t.Run("empty payload")` — construir token sin payload field.

**Tests para `NewOpaqueMethod` (en `tokens_test.go`):**
- `TestNewOpaqueMethod` con subtests:
  - `t.Run("default options")` — verificar que cipher, generateFn, validateFn estan seteados.
  - `t.Run("with custom key")` — pasar `WithKey`, verificar roundtrip.
  - `t.Run("with issuer")` — pasar `WithIssuer`, verificar que generate incluye issuer.
  - `t.Run("with timeout")` — pasar `WithTimeout`, verificar expiracion.

**Tests para predefined methods (en `tokens_test.go`):**
- `TestOpaqueAES256GCM` — roundtrip con `OPAQUE_AES_256_GCM`.
- `TestOpaqueXChaCha20Poly1305` — roundtrip con `OPAQUE_XCHACHA20_POLY1305`.

**Tests de registry (en `extensions_test.go`):**
- Verificar que `Get("OPAQUE_AES_256_GCM")` retorna el method correcto.
- Verificar que `Supported()` incluye los opaque methods.

**Reglas de testing (del CODING_STANDARDS):**
- `t.Parallel()` en todos los tests y subtests.
- `t.Fatal`/`t.Fatalf`, no testify.
- No table-driven.
- 100% statement coverage (exceptuando branches defensivos de crypto stdlib).
- Override: no parallel en `extensions_test.go` (estado global del registry).

##### Paso 10 — Actualizar examples

Actualizar `tokens/examples/main.go` para demostrar:
1. Uso directo de `OPAQUE_AES_256_GCM` — generate + validate.
2. Uso de `NewOpaqueMethod` con key y cipher custom.
3. Registry lookup con `Get("OPAQUE_AES_256_GCM")`.
4. `Supported()` mostrando JWT y opaque methods juntos.

##### Paso 11 — Linter y coverage

1. `go tool golangci-lint run --fix ./modules/common/crypto/tokens/...` — 0 issues.
2. Coverage >= 97% (branches defensivos de crypto stdlib se exceptuan).
3. Verificar que `.testcoverage.yml` no necesita cambios (examples ya estan excluidos).

##### Paso 12 — go.mod

Verificar que `modules/common/go.mod` ya tiene la dependencia de `ciphers/aead` (es del
mismo modulo `common`, asi que no hay que agregar nada externo). No se agregan dependencias nuevas.

#### Archivos modificados

| Archivo | Cambio |
|---------|--------|
| `tokens/tokens.go` | Nuevo campo `cipher`, `NewOpaqueMethod`, predefined opaque vars |
| `tokens/types.go` | Package doc actualizado, type compliance para `generateOpaque`/`validateOpaque` |
| `tokens/functions.go` | Nuevas funciones `generateOpaque`, `validateOpaque` |
| `tokens/options.go` | Nuevo campo `cipher` en `Options`, nueva `WithCipher` option |
| `tokens/errors.go` | Nuevos sentinel errors para opaque |
| `tokens/extensions.go` | Agregar opaque methods al registry init |
| `tokens/tokens_test.go` | Tests para `NewOpaqueMethod` y predefined methods |
| `tokens/functions_test.go` | Tests para `generateOpaque` y `validateOpaque` |
| `tokens/extensions_test.go` | Tests de registry para opaque methods |
| `tokens/examples/main.go` | Demos de opaque token usage |

#### Dependencias internas

```
tokens
  ├── common/crypto/ciphers/aead   (nuevo import)
  ├── common/assert
  ├── common/errs
  ├── common/random
  ├── common/utils
  └── github.com/golang-jwt/jwt/v5 (existente, solo para JWT methods)
```

#### Lo que NO cambia

- La API publica existente de JWT no se modifica.
- `Method.Generate` y `Method.Validate` siguen delegando a `generateFn`/`validateFn` — el dispatch entre JWT y opaque es por la funcion inyectada, no por if/else.
- Los tests existentes de JWT no se tocan.
- El registry es backward-compatible: `Get("JWT_HS256")` sigue funcionando igual.

---

### DelegatingPasswordEncoder — `common/crypto/passwords`

Agregar un password encoder que routea por prefijo. Migrado de go-feather-lib `security/`.

**Concepto:**
Passwords almacenados con prefijo de algoritmo: `{bcrypt}$2a$10$...`, `{argon2}$argon2id$...`.
El delegating encoder detecta el prefijo y delega al encoder correcto para verificar.
Para encoding nuevo, usa un encoder default configurable.

**Dependencia:** Solo `common/crypto/passwords` (los encoders individuales ya existen).

---

## observability

### HTTP RoundTripper decorators — `telemetry/` o `common/http`

Decorators de `http.RoundTripper` para instrumentacion automatica de HTTP clients:

- **Metrics RoundTripper** — request count (method, host, path, status) + duration histogram (Prometheus).
- **Tracing RoundTripper** — OTel spans por request, inject W3C trace context headers, record timing/status/errors.

Ambos se stackean: `transport -> tracing -> metrics -> base`.

**Dependencia:** telemetry/otel (OTel SDK), prometheus client.

---

## resilience

### Circuit Breaker & Rate Limiter — `common/resilience/`

Registries thread-safe para patrones de resiliencia en llamadas outbound:

- **CircuitBreakerRegistry** — lazy-create por nombre, usa `sony/gobreaker`. Defaults: 3 max requests, 60s interval, 15s timeout, 5 consecutive failures to trip.
- **RateLimiterRegistry** — lazy-create por nombre, usa `golang.org/x/time/rate`. Defaults: 100ms rate, burst 5.

Patron: registry con mutex, `Get(name)` crea si no existe con defaults, configurable via options.

**Dependencia:** `sony/gobreaker`, `golang.org/x/time/rate`.

---

## boot

### Application Wiring — nuevo modulo `modules/boot/` o extension de `modules/config/`

Framework de wiring que conecta config (bootstrap) con managed (lifecycle).

**Problema que resuelve:**
Hoy yarumo tiene config (bootstrap one-shot) y managed (lifecycle start/stop/done),
pero no hay un mecanismo formal para conectar componentes entre si. El wiring se hace
manual en `sample/main.go`. Esto es el eslabon faltante.

**go-feather-lib tenia `boot/` con:**
- `ApplicationContext` — god-struct con ~30 campos: app metadata, environment, database (GORM),
  security (password encoder/generator/manager, token manager, auth service/filter, authz
  service/filter), HTTP (gin router), gRPC (service desc + server). Monolitico y acoplado
  a gin, GORM, grpc directamente.
- `BeanBuilder` — struct con 17 factory functions, una por componente. Cada una recibe
  `*ApplicationContext` y retorna el componente construido. Defaults incluidos (bcrypt,
  JWT, gin routes /login /health /info /api).
- `Init()` — crea `ApplicationContext` via `NewApplicationContext()`, llama delegate function,
  attacha servers a lifecycle (`qmdx00/lifecycle`), ejecuta `app.Run()`.
- `Enablers` — feature toggles (HttpServerEnabled, GrpcServerEnabled, DatabaseEnabled).
- Orden secuencial fijo: environment -> config -> datasource -> security -> http -> grpc.

**Problemas del diseno original:**
1. Acoplamiento directo a gin, GORM, grpc — si no usas uno, igual lo importas.
2. God-struct — `ApplicationContext` con 30+ campos publicos.
3. Orden fijo — no se puede cambiar la secuencia de inicializacion.
4. Sin generics — `any` para gRPC service server.

**Sandbox (eliminado) tenia una version mejorada:**
- `Container` con campos tipados + `more map[string]any` extensible.
- `WireContext[C any]` generico — embeds Container + typed Config.
- `BeanFn func(*Container)` — simples factory functions.
- `Run[C]()` — integra signal handling + managed lifecycle.
- Menos acoplado pero sin diseño formal.

**Diseno sugerido para yarumo:**

```go
// Container holds wired components. Extensible via typed map.
type Container struct {
    components map[reflect.Type]any
}

// Register[T] registers a component by type.
func Register[T any](c *Container, component T)

// Resolve[T] retrieves a component by type.
func Resolve[T any](c *Container) (T, error)

// BeanFn is a factory that populates the container.
type BeanFn func(ctx context.Context, c *Container) error

// Run orchestrates: config.Default -> BeanFns -> managed lifecycle -> signal wait.
func Run(ctx context.Context, name, version, env string, beans ...BeanFn) error
```

**Dependencia:** config, managed, common/crypto/*.

Requiere disenar cuidadosamente para no acoplar. El container debe ser generico y no
conocer tipos especificos (no importar gin, GORM, etc.).

---

## Migracion go-feather-lib — Estado

### Ya migrado

| go-feather-lib | yarumo | Estado |
|---|---|---|
| `common/assert` | `common/assert` | Completo |
| `common/constraints` | `common/constraints` | Completo |
| `common/errors` | `common/errs` | Mejorado (TypedError, ErrorInfo) |
| `common/http` | `common/http` | Mejorado (Client con rate limiting, retry) |
| `common/log` | `common/log` + `log/slog` | Mejorado |
| `common/rest` | `common/rest` | Rediseñado |
| `common/ssl` | `common/crypto/certs` | Migrado y expandido |
| `common/utils` | `common/utils` | Expandido (Filter, Map, Chunk, Deduplicate) |
| `common/server` | `modules/managed` | Rediseñado como lifecycle components |
| `uxid` | `common/uids` | Completo |
| `security/passwords` | `common/crypto/passwords` | Migrado (argon2, bcrypt, pbkdf2, scrypt) |
| `security/tokens` | `common/crypto/tokens` | Migrado (JWT HMAC) |
| `boot` (parcial) | `modules/config` | Solo bootstrap. Wiring pendiente (ver seccion boot) |
| `common/config` | `modules/config` | Absorbido por viper |
| `common/environment` | `modules/config` | Absorbido por viper |
| `common/properties` | `modules/config` | Absorbido por viper |

### No migrado — pendiente (ver BRAINSTORM.md para detalles)

| go-feather-lib | Prioridad | Destino |
|---|---|---|
| `security/DelegatingPasswordEncoder` | Alta | `common/crypto/passwords` |
| `boot/` (wiring + DI) | Alta | `modules/boot/` (ver seccion boot arriba) |
| `security/AuthenticationService` | Media | `modules/auth/` |
| `security/AuthorizationFilter` | Media | `modules/auth/` |
| `security/PrincipalManager` | Media | `modules/auth/` (depende de datasource) |
| `health/` | Media | `common/health/` o `managed/` |
| `datasource/gorm` | Media | `modules/datasource/` |
| `datasource/mongo` | Baja | `modules/datasource/` |
| `datasource/goredis` | Baja | `modules/datasource/` |
| `datasource/gocql` | Baja | `modules/datasource/` |
| `integration/messaging/` (EIP) | Baja | `modules/messaging/` (capa de abstraccion) |
| `messaging/rabbitmq/amqp` | Baja | `modules/messaging/rabbitmq/amqp/` |
| `messaging/rabbitmq/streams` | Baja | `modules/messaging/rabbitmq/streams/` |
| `common/stats` | Baja | `maths/stats/` |
| `common/validation` | Baja | Evaluar vs go-playground/validator |

### Descartado / no aplica

| go-feather-lib | Razon |
|---|---|
| `common/collections` | Cubierto por `common/utils` + `maths/sets` |
| `web/` | Cubierto por `common/log` |
| `cache/` | Directorio vacio, nunca implementado. Si se necesita, disenar desde cero (ver BRAINSTORM.md) |
| `messaging/kafka/` | Directorio vacio en go-feather-lib |
| `messaging/nats/` | Directorio vacio en go-feather-lib |
