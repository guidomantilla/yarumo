# PACKAGES

## Paquetes-librería sin estado

### Cómo se reconocen

- API solo de funciones libres. Mismos args ⇒ mismo resultado.
- Ningún `New<X>(...)`, ningún tipo con métodos de negocio.
- Ejemplos en el repo: `common/assert/`, `common/cast/`, `common/utils/`, `common/pointer/`, `common/random/`, `common/validation/`.

### Inventario en `modules/common/`

| Paquete | Qué hace |
|---|---|
| `assert/` | Assertions runtime (`NotNil`, `NotEmpty`, `Equal`, `True`, `False`) — modo log o fatal según config. |
| `cast/` | Type-safe casting (`ToInt`, `ToString`, `ToTime`, `ToDuration`, …) — wrappa `spf13/cast`. |
| `crypto/random/` | Generación crypto-segura de bytes, números y strings. |
| `errs/` | Typed errors + error-chain helpers (`As`, `Match`, `Wrap`, `Unwrap`, `ErrorMessages`, `AsErrorInfo`) + JSON-serializable info. |
| `pointer/` | Helpers para pointers (deref con default, take-address, comparación). |
| `random/` | Generación pseudoaleatoria no-crypto (`Bytes`, `Number`, `String`, `Text*`). |
| `utils/` | Funciones genéricas (`Coalesce`, `Ternary`, `Empty`, `RandomString`, case-helpers `PascalCase`/`SnakeCase`/…). |
| `validation/` | Leaves de validación (`IsRequired`, `MinLen`, `IsEmail`, `IsUUID`, …) + reflexión por dotted path (`GetField`). |


## Reglas

### 1. Layout de archivos

- **Trío base**: `types.go` + `functions.go` + `functions_test.go`.
- **Partición por visibilidad**:
  - `functions.go` → todo lo público (funciones exportadas).
  - `internals.go` → todo lo privado (funciones helpers no exportados).
- **Excepción**: un struct privado cuyo único propósito es soportar a los helpers de `internals.go` (ej. `pathSegment` en `validation/internals.go`, que parametriza a `parsePath`/`walkSegment`/etc.) puede convivir con ellos en el mismo archivo. No aplica a types privados de otra naturaleza.
- **Opcionales según necesidad del paquete**:
  - `constants.go` — constantes públicas o privadas que no sean sentinel-errors.
  - `options.go` / `options_test.go` — si alguna pública toma el patrón Options.
  - `errors.go` — si el paquete declara errores de dominio.

### 2. `types.go`

- Un `Fn` alias por **cada** función pública (`type MinLenFn func(s string, n int) error`).
- Bloque `var (_ XxxFn = Xxx ...)` con compliance **exhaustiva** sobre todas las públicas → compile-time typecheck del contrato firma↔alias.
- Para genéricos, instanciar con un tipo representativo (`_ CheckFn[string] = IsEmail`, `_ MinFn[int] = Min[int]`). Cualquier cambio incompatible de firma rompe igual en compile-time.
- **Excepción**: `NewOptions(opts ...Option) *Options` del Options pattern no necesita `Fn` alias ni compliance. El contrato lo aporta el type `Option func(*Options)`, que ya enforza signature en cada `WithXxx`.

### 3. Errores de dominio (cuando aplica)

- Viven en `errors.go`:
  - Sentinels con `errors.New(...)`.
  - `Error` struct embebiendo `cerrs.TypedError` con `Type` constante.
  - Factory `ErrXxx(causes ...error) error` que envuelve causas y `ErrXxxFailed`.

### 4. Documentación

Las reglas universales del repo (doc terminado en punto, comenzar por el nombre del identificador) están en `CODING_STANDARDS.md`. Las específicas del shape:

- **Package doc** vive en `types.go` (único archivo del trío base que siempre está): una frase qué provee, opcionalmente un segundo párrafo de contrato/integración.
- **`Fn` aliases**: una línea — `// XxxFn is the function type for Xxx.`
- **Funciones (públicas y privadas)**: describen resultado o condición de error, no implementación. Las privadas de `internals.go` con el mismo rigor que las públicas.
- **`constants.go`**: comentario de grupo encima del bloque `const (...)`. Sin doc por constante salvo que el nombre no sea autoexplicativo.
- **`options.go`**: cada identificador (`Options`, `NewOptions`, `Option`, cada `WithXxx`) con doc dedicado. Modelo de referencia: `modules/common/utils/options.go`.
- **`errors.go`**:
  - Sentinels: comentario de grupo encima del bloque `var (...)`; cada sentinel se autodocumenta con su mensaje.
  - `Error` struct + factory `ErrXxx`: doc-comment dedicado a cada uno.

**Referencia operativa**: `modules/common/utils/` y `modules/common/validation/` cubren entre los dos todos los casos. Si dudás cómo documentar algo, mirá esos paquetes.

