# PACKAGES

## Paquetes-librería sin estado

### Cómo se reconocen

- API expone funciones libres como superficie principal. El consumidor llama funciones del paquete, no construye y opera sobre tipos del paquete.
- No exporta structs con invariantes mutables ni constructores `NewXxx(opts ...Option) Interface` con Options pattern. Pueden existir constructores triviales que devuelven valores inmutables bajo interface (ej. `NewUID(name, fn) UID`) sin descalificar el shape.
- Pueden mantener **estado mutable interno** detrás de las funciones libres (PRNG state, slot `current` swappable vía `Use`, registry map, regex caches). La pureza "mismos args ⇒ mismo resultado" es ideal pero no requisito: ver bloque "Estado mutable de paquete" al final de esta sección.
- Ejemplos en el repo: `common/assert/`, `extension/common/cast/`, `common/utils/`, `common/pointer/`, `common/random/`, `extension/common/validation/`, `common/log/`, `extension/common/uids/`.

### Inventario en `modules/core/common/`

| Paquete | Qué hace |
|---|---|
| `assert/` | Assertions runtime (`NotNil`, `NotEmpty`, `Equal`, `True`, `False`) — modo log o fatal según config. |
| `errs/` | Typed errors + error-chain helpers (`As`, `Match`, `Wrap`, `Unwrap`, `ErrorMessages`, `AsErrorInfo`) + JSON-serializable info. |
| `log/` | Facade abstracta de logging estructurado (`Logger` interface + `Use`/`Default` + helpers `Trace`/`Debug`/`Info`/`Warn`/`Error`/`Fatal`) sobre slot mutable. Trío base (`types.go`/`functions.go`/`functions_test.go`) + `internals.go` con `loggerHolder` (struct sin métodos, excepción del consumidor de `load`) + vars `current`/`internal`/`osExit` + helper `load`. Concern del default `noopLogger` (struct privado con métodos que implementa `Logger`) aislado en `noop.go`. Implementaciones concretas viven en `modules/extension/common/log/`; este paquete no depende de ninguna. Default noopLogger (Fatal escribe a stderr y exit) hasta que el consumer llame `Use(...)`. |
| `pointer/` | Helpers para pointers (deref con default, take-address, comparación). |
| `random/` | Generación pseudoaleatoria no-crypto (`Bytes`, `Number`, `String`, `Text*`). |
| `rest/` | Cliente REST stateless (`Call`, `CallStream`, `DecodeHTTPError`) con DTOs `RequestSpec`/`ResponseSpec[T]`/`StreamResponseSpec`; concurrency-safe. |
| `utils/` | Funciones genéricas (`Coalesce`, `Ternary`, `Empty`, `RandomString`, case-helpers `PascalCase`/`SnakeCase`/…). |

### Extensiones a las reglas

Aplican las 4 reglas universales con las siguientes extensiones cuando el paquete necesita exponer concerns con tipos asociados además de funciones libres:

- **R1 (Concern por archivo)** — un paquete puede agrupar un concern en su propio archivo, nombrado por el concern (ej. `specs.go`, `uids.go`, `extensions.go`). El archivo de concern reúne **todo** lo que pertenece a ese concern: struct (público o privado), constructores, métodos, singletons preconfigurados, vars privadas de estado, regexes/constantes privadas. La partición visibility (functions.go/internals.go) **no aplica dentro de un archivo de concern**: públicos y privados conviven porque la unidad es el concern.
  - **Variante DTO público con métodos puros**: ej. `rest/specs.go` — `RequestSpec`/`ResponseSpec` públicos con `Build` (pública) + 3 helpers privados (`buildHeaders`, `marshalBody`, `buildURL`).
  - **Variante struct privado expuesto vía interface pública**: ej. `uids/uids.go` — struct `uid` privado, interface `UID` pública, `NewUID` (constructor trivial) + métodos + singletons preconfigurados (`UuidV4`, `NanoID`, …).
  - **Variante funciones libres de un concern ajeno al resto del paquete**: ej. `uids/extensions.go` — registry global del paquete con state (`methods` map, `current`, `lock`) + funciones de registro (`Register`/`Get`/`Use`/`Generate`/`Supported`). Las funciones públicas no se mueven a `functions.go` porque pertenecen al concern del archivo, no al pool general de funciones libres.
- **R2 (Métodos sobre tipos del concern)** — los métodos (públicos o privados) sobre tipos del concern no necesitan `Fn` alias ni compliance. Su signatura está atada al receiver type; cualquier drift cae en compile-time vía consumers. Aplica también a métodos sobre `Error` struct en `errors.go`.
- **R4 (Métodos sobre tipos del concern)** — cada método (público o privado) sobre un tipo del concern lleva su doc-comment, mismo rigor que las funciones libres.

**Estado mutable de paquete (singleton interno) no descalifica de Shape A.** Un paquete sigue siendo Shape A cuando su API son funciones libres aunque internamente mantenga estado (`current` slot swappable vía `Use`, registry map, PRNG state, regex caches). Las reglas "mismos args ⇒ mismo resultado" y "ningún `New<X>(...)`" son guías sobre la **forma de la API pública**, no prohibiciones sobre la implementación. Precedentes: `random/` (PRNG interno, `randInt` mockeable), `common/log/` (slot `current atomic.Value` swappable vía `Use(Logger)`), `uids/` (registry global swappable vía `Use(name)`). Un constructor trivial que devuelve un valor inmutable bajo interface (ej. `NewUID(name, fn) UID`) tampoco descalifica — no hay invariantes mutables ni Options pattern.


## Paquetes-librería con estado

### Cómo se reconocen

- Expone tipos con estado: structs que mantienen invariantes entre llamadas (claves, contadores, buffers, conexiones, configuración) y se operan a través de métodos.
- El API entra por uno o más constructores `NewXxx(opts ...Option) Interface` que devuelven una **interface** declarada en `types.go`. El struct concreto la implementa.

### Inventario en `modules/core/common/`

| Paquete | Qué hace |
|---|---|
| `cache/` | `Cache[K, V]` genérico embebiendo `lifecycle.Component` + backend de referencia in-memory (`NewMemoryCache`) + primitivas compartidas `Codec`/`JSONCodec`/`ResolveKeyPrefix`. |
| `expressions/` | `Evaluator` de expresiones — lexer/parser/eval sobre AST con scope. |
| `health/` | Aggregator de health checks que orquesta múltiples sondas. |
| `http/` | `Client` + `Server` HTTP con retry/limiter, defaults seguros para timeouts/headers. |

**Concerns ajenos en archivos `<concern>.go`** — mismo principio que R1 de Shape A (concern por archivo) aplicado a funciones libres: cuando dos grupos de funciones libres pertenecen a concerns ajenos entre sí, se separan en archivos `<concern>.go` en vez de mezclarse en `functions.go`. Precedentes: `uids/extensions.go` (registry global del paquete con state + `Register`/`Get`/`Use`/`Generate`/`Supported`, distinto de los generadores/validadores libres en `functions.go`); `diagnostics/handlers.go` (HTTP handlers `NewPprofHandler` distintos de las capturas de profile en `functions.go`) — ahora en `modules/managed/diagnostics/` tras el split de #175.

### Diferencias con las reglas

Aplican las 4 reglas universales del documento con las siguientes extensiones:

- **R1 (Layout)** — además del trío base, cada struct stateful tiene su **propio archivo nombrado como el struct** (minúsculas: `client.go`, `server.go`). Ese archivo contiene declaración del struct + constructor `NewXxx` + métodos. `functions.go` sigue siendo para funciones libres públicas (típicamente helpers o defaults usados para configurar Options).
- **R1 (Naming de structs multi-palabra)** — si el struct tiene nombre CamelCase compuesto:
  - Implementación única o canónica de una interface → `<nombre>.go` con todo en minúsculas y sin separadores (ej. `CircuitBreaker` → `circuitbreaker.go`).
  - Variante alternativa de una interface que ya tiene archivo canónico → `<canonical>_<variante>.go`, agrupando por la interface (ej. `PluggableClient`, variante de la implementación de `Client`, vive en `client_pluggable.go`).
  - **Múltiples peers de una interface sin canónica** — cuando el paquete expone N implementaciones equivalentes de la misma interface (típicamente stdlib o externa, como `slog.Handler`) y ninguna es "la canónica", se nombran `<role>_<variante>.go` donde `<role>` es el role compartido en minúsculas y `<variante>` distingue cada impl. Ej.: `fanoutHandler` y `contextHandler` ambos implementan `slog.Handler` y viven en `handler_fanout.go` + `handler_context.go`.
- **R1 (Singletons preconfigurados)** — instancias `var` construidas en tiempo de carga del paquete (ej. `DefaultClient`, `NoopClient`, `ErrorClient`) van al inicio del archivo de la implementación canónica de su interface, incluso si algún singleton concreto se construye literalmente con una variante (ej. `NoopClient = &PluggableClient{…}` vive en `client.go`, no en `client_pluggable.go`, porque pertenece a la familia "Client").
- **R1 (Tipos privados compartidos)** — cuando dos o más structs stateful del paquete comparten un tipo privado (struct de datos, enum o ambos) que no es helper de `internals.go` sino estructura común consumida por varios concerns, ese tipo vive en su propio archivo `<concern>.go` junto con sus constantes asociadas. Ej.: `expressions/tokens.go` con `token` struct + `tokenKind` enum + sus constantes `tokEOF`/`tokPlus`/etc., compartidos entre `lexer` y `parser`. No aplica si el tipo privado es usado por un solo consumidor (en ese caso vive con su consumidor o en `internals.go`).
- **R2 (`types.go`)** — además de los `Fn` aliases, declara las **interfaces** del paquete. Cada implementación tiene su compliance var: `var _ Interface = (*impl)(nil)`.
- **R2 (Excepción adicional)** — los constructores `NewXxx(opts ...Option) Interface` no necesitan `Fn` alias ni compliance. El contrato ya está fijado por el Options pattern en la entrada y por la interface + su compliance en la salida.
- **R4 (Interfaces)** — cada interface en `types.go` lleva doc-comment que enuncia el contrato: propósito + expectativas de concurrencia + responsabilidad del caller (cleanup, lifecycle, cancelación). Cada método de la interface lleva su propio doc-comment.
- **R4 (Métodos)** — la regla "funciones (públicas y privadas) llevan doc" se extiende a métodos: cada método (público o privado) sobre un struct stateful lleva doc-comment con el mismo rigor.
- **R4 (Singletons preconfigurados)** — el bloque `var (...)` de singletons lleva un comentario de grupo encima describiendo la familia, mismo patrón que sentinels de errores.


## Reglas

### 1. Layout de archivos

- **Trío base**: `types.go` + `functions.go` + `functions_test.go`.
- **Partición por visibilidad** (aplica **solo a funciones libres**, no a métodos sobre tipos ni a vars/structs/constantes asociadas a un concern):
  - `functions.go` → funciones libres públicas (exportadas).
  - `internals.go` → **exclusivamente funciones libres privadas** (helpers no exportados, consumidos por las funciones públicas). Su test file gemelo es `internals_test.go`. **No es vertedero de "todo lo privado"**: vars privadas de paquete (`current`, `methods`, regexes, mutex), structs privados con métodos, y constantes privadas asociadas a un concern viven con su consumidor (archivo de concern correspondiente o el archivo de la función pública que las usa). Ver R1 de Shape A para archivos de concern.
- **Test del archivo `internals.go`: opcional.** La cobertura vía API público suele bastar. Solo crear `internals_test.go` (white-box) cuando un helper sea lo suficientemente complejo o independiente como para que probar transitivamente oculte gaps de cobertura.
- **Todos los archivos `_test.go` usan `package <name>` (internal).** El sufijo `_test` en el package (`package <name>_test`) **no está permitido**. La separación external/internal complica el layout sin beneficio claro y obliga a tener dos archivos por source. Un único test file por source, en el mismo package, cubre todos los casos. Aplica también a directorios test-only (ej. `compute/tests/acceptance/` usa `package acceptance`, no `acceptance_test`).
- **Si `functions.go` queda vacío** (todas sus privadas se movieron a `internals.go` y no hay públicas libres) → no debe existir. La regla del trío base no obliga a mantener un archivo vacío.
- **Si no hay funciones privadas helpers**, `internals.go` no existe. No se crea para albergar vars o structs privados — esos van con su consumidor.
- **Excepción al alcance de `internals.go`**: un struct privado cuyo único propósito es soportar a los helpers de `internals.go` (ej. `pathSegment` en `validation/internals.go`, que parametriza a `parsePath`/`walkSegment`/etc.) puede convivir con ellos en el mismo archivo. **No aplica a types privados de otra naturaleza** — si el struct privado tiene métodos que satisfacen una interface pública del paquete, va a su propio archivo de concern (ver R1 de Shape A).
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
- **Structs (públicos y privados)**: cada `type Xxx struct` lleva doc-comment de 1-2 líneas describiendo qué representa o qué papel cumple (ej. "client implements Client.", "serviceRegistration carries a service impl + its descriptor for late registration."). Aplica también a los structs auxiliares dentro de `options.go` y a los structs implementadores en archivos per-tipo de Shape B.
- **`constants.go`**: comentario de grupo encima del bloque `const (...)`. Sin doc por constante salvo que el nombre no sea autoexplicativo.
- **`options.go`**: cada identificador (`Options`, `NewOptions`, `Option`, cada `WithXxx`) con doc dedicado. Modelo de referencia: `modules/core/common/utils/options.go`.
- **`errors.go`**:
  - Sentinels: comentario de grupo encima del bloque `var (...)`; cada sentinel se autodocumenta con su mensaje.
  - `Error` struct + factory `ErrXxx`: doc-comment dedicado a cada uno.

**Referencia operativa**: `modules/core/common/utils/` y `modules/extension/common/validation/` cubren entre los dos todos los casos. Si dudás cómo documentar algo, mirá esos paquetes.

## Excepciones a los shapes

Algunos paquetes bajo `common/` no encajan en ningún shape — porque son envoltorios delgados sobre una librería externa, tienen un constraint de dependencias que justifica la desviación, o su superficie es exclusivamente declaración de tipos (sin funciones libres como API principal). Quedan fuera del inventario de Shape A y Shape B, y de sus reglas.

- `modules/extension/common/log/slog/` — adapter sobre `log/slog` stdlib que **extiende** el tipo con métodos propios (`Trace`, `Fatal`). Expone `Logger` como **struct público concreto** (no como interface) e implementa la interface `common/log.Logger` (typing estructural). Vive como módulo top-level porque depende de `common/log` (interface) en dirección consumer → abstracción; el ciclo arquitectónico inverso (common → impl) queda eliminado. Esta forma encaja en la excepción 4 de `CODING_STANDARDS.md` (criterio 4) y rompe también el patrón Shape B clásico, así que vive acá.
- `common/constraints/` — solo declara type constraints genéricas (`Signed`, `Unsigned`, `Integer`, `Float`, `Complex`, `Number`) + aliases (`Comparable`, `Ordenable`). Sin funciones libres, sin métodos, sin estado. Análogo a `golang.org/x/exp/constraints`. No tiene `functions.go` (no hay funciones); el package doc + declaraciones viven en `types.go` (único archivo).
- `common/types/` — solo declara el tipo `Bytes []byte` con métodos puros (`ToHex`, `ToBase64Std`/`ToBase64RawStd`/`ToBase64Url`/`ToBase64RawUrl`). Sin funciones libres del paquete. Encaja parcialmente en R1 variante 1 de Shape A (DTO público con métodos puros), pero no cumple el trío base porque no hay funciones libres que justifiquen `functions.go` ni un `types.go` separado del concern: el package doc + tipo + métodos viven todos en `bytes.go` (único archivo).
- **Subpaquetes de `modules/core/crypto/`** (14 paquetes — ver inventario abajo) — siguen el **Crypto Subpackage Standard** documentado en `modules/core/crypto/CODING_STANDARDS.md`. El standard define file structure propia (`types.go`, `errors.go`, `<name>.go`, `functions.go`, `options.go`, `extensions.go`, `text_codec.go`) y overrides explícitos a 3 criterios del documento general: criterion 3 (struct público concreto, no interface), criterion 4 (constructor devuelve `*Method` con pluggable function fields), criterion 6 (registry multi-instance, no singleton `Use`). Aplica al universo crypto completo, con 3 utility packages (`random/`, `certs/`, `passwords/generator/`) que no usan el Method pattern y siguen Shape A. Para detalles y compliance ver el standard; PACKAGES.md no duplica esas reglas.
- **`common/lifecycle/`** — única excepción al principio "common no tiene lifecycle ni dispara goroutines". El paquete **es la primitiva del lifecycle del workspace**: declara la interface `Component` (Name/Start/Stop/Done), los helpers `Start`/`Stop` que coordinan el run lifecycle, los tipos `ErrChan`/`CloseFn`, y los `Build*` builders canónicos que componen un `Component` con su goroutine de arranque y su `CloseFn` de teardown. Está permitido — y esperado — que sus funciones disparen goroutines (`go component.Start(...)` dentro de `Build*`) y que importen `common/log` para emitir las líneas de boundary (`starting up` / `stopping` / `stopped` / `failed to start` / `shutdown failed`). El resto de `common/` sigue sin lifecycle. Este es el único punto del subsistema común autorizado a manejar concurrencia activa; consumers que necesiten un lifecycle propio extienden la interface, no replican la maquinaria. Code reviews deben tratar cualquier `go ...` o `Build*` fuera de `common/lifecycle/` como red flag — pertenece a un módulo top-level (`http/`, `cron/`, `grpc/`, `diagnostics/`, etc.), no a `common/`.

## Módulo `modules/core/crypto/`

Módulo top-level extraído de `modules/core/common/crypto/` (issue #170). Reúne 14 subpaquetes con un único `go.mod` y un standard propio (`modules/core/crypto/CODING_STANDARDS.md`).

### Inventario

| Subpaquete | Tipo | Qué hace |
|---|---|---|
| `random/` | Utility (Shape A) | Generación crypto-segura de bytes, números y strings. |
| `certs/` | Utility | Helpers TLS/x509 (CSR, self-signed certs, pool builders, PEM I/O). |
| `passwords/generator/` | Utility (`*Generator`) | Constructor de passwords aleatorios configurable (longitud, charset, política). |
| `hashes/` | Method | Hash funcs registradas (SHA-2/3, BLAKE2, …) + `Hash`/`Compute` libres. |
| `kdfs/` | Method | Key derivation functions (HKDF, PBKDF2, scrypt, argon2id). |
| `passwords/` | Method | Password hashing/verification (bcrypt, scrypt, argon2, pbkdf2) + delegating encoder. |
| `tokens/` | Method | JWT signing/verification sobre `golang-jwt/jwt/v5` con algoritmos registrables. |
| `ciphers/aead/` | Method | AEAD (AES-GCM, ChaCha20-Poly1305) + streaming. |
| `ciphers/hybrid/` | Method | HPKE (X25519 + AEAD vía circl). |
| `ciphers/rsaoaep/` | Method | RSA-OAEP encryption. |
| `signers/ecdsas/` | Method | ECDSA con curvas P-256/384/521. |
| `signers/ed25519/` | Method | Ed25519. |
| `signers/hmacs/` | Method | HMAC (SHA-256/384/512). |
| `signers/rsassas/` | Method | RSASSA (PKCS#1 v1.5 + PSS, SHA-256/384/512). |

## Contenedor `modules/extension/common/log/`

Directorio puramente contenedor — **NO es un go module**. Aloja **módulos hermanos**, uno por implementación concreta de la interface `common/log.Logger`. Cada impl tiene su propio `go.mod` y se promociona/retira independientemente. La abstracción (interface, slot global, helpers `Trace`/.../`Fatal`) vive en `common/log/`; este directorio aporta las impls.

**Dirección de dependencia.** `modules/extension/common/log/<impl>` → `common/log` (interface). Nunca al revés. Esta inversión es lo que mantiene `common/` libre de dependencias hacia módulos top-level y evita el ciclo arquitectónico `common → log → common`.

| Submódulo | Shape | Qué hace |
|---|---|---|
| `log/slog/` | Excepción (struct público concreto, no interface) | Adapter sobre `log/slog` stdlib que **extiende** el tipo con métodos propios (`Trace`, `Fatal`). Expone `*Logger` como struct público concreto e implementa `common/log.Logger` por typing estructural. Incluye `Options` (`WithLevel`/`WithWriter`/`WithHandlers`/`WithContextExtractors`), `NewFanoutHandler`, `NewContextHandler`, `ReplaceLevel`, `SlogctxExtractor`. Aloja el subpaquete `slogctx/` (subpaquete propio dentro del mismo módulo, sin go.mod separado). |
| `log/slog/slogctx/` | Shape A | Subpaquete del módulo `slog/`. Bag context-bound de `slog.Attr` (`WithAttrs`, `SetAttrs`, `Attrs`) para propagar atributos por `context.Context`. Sin estado de paquete. |
| `log/zerolog/` | Shape B | Adapter sobre `github.com/rs/zerolog`. `NewLogger(opts ...Option) clog.Logger` retorna la interface (patrón canónico; slog es la excepción histórica). Options: `WithLevel`/`WithWriter`/`WithConsole`/`WithTimeFormat`/`WithSampling`. Args variádicos se parsean como pares clave-valor con dispatch tipado en `internals.go`. Sin registry, sin pluggable. |

Los tests de la facade `common/log/` son intencionalmente seriales (sin `t.Parallel()`) porque mutan el slot global. Documentado en cabecera de `common/log/functions_test.go` y en `common/log/doc.go`. Los submódulos de `modules/extension/common/log/` (`slog/`, `slog/slogctx/`, `zerolog/`) corren con `t.Parallel()` en todos sus tests.

Histórico: `common/log/` fue extraído como módulo top-level en #173, pero esto cerró un ciclo arquitectónico con `common/assert` que dependía de log. La reorganización en ese PR devolvió la **interface** a `common/log/` y dejó en `modules/extension/common/log/` solo las **implementaciones concretas** — patrón paralelo a `commons-logging`/`slf4j-api` vs binding impls. Tras la entrada de `zerolog/` como módulo separado, el directorio `extension/common/log/` se reestructuró: el `go.mod` parent se eliminó y `slog/` se promovió a módulo hermano, restaurando la simetría (`slog/` y `zerolog/` son peers, cada uno con go.mod propio).

## Módulo `modules/messaging/`

Top-level module que aloja **dos capas** de la stack de mensajería in-process: el **paquete root** con las primitivas (`Channel[T]`, `Message[T]`, las 5 implementaciones de canal), y los **sub-paquetes** con patrones EIP que componen sobre esas primitivas (router, filter, bridge — más patterns por venir). Cada sub-paquete importa el root package; el root no importa nada de sub-paquetes (sin ciclos). Comparten `go.mod` — el módulo entero arrastra solo `core/common` y stdlib, sin deps externas pesadas. Transports a brokers externos (Kafka, NATS, etc.) viven en `extension/messaging/<broker>/` cuando se implementen.

### Capa 1: Primitives (root del módulo)

Expone una sola interface pública `Channel[T]` con cinco implementaciones equivalentes que cubren los ejes ortogonales sync/async × fan-out/point-to-point + sink. Sigue **Shape B con múltiples peers de una interface sin canónica** — file naming `<variant>.go`, structs `<variant>` per R1.

| Variante | Archivo | Struct | Roles | Qué hace |
|---|---|---|---|---|
| Pipeline | `pipeline.go` | `pipeline[T]` | `Channel[T]` | Sync sequential fan-out en la goroutine del caller. Fail-fast con `*ChainError` trace (step por step: ok / error / panic / skipped). Patrón "Transactional Handler Chain" (cf. MediatR pipeline behaviors). |
| Broadcast | `broadcast.go` | `broadcast[T]` | `Channel[T]` | Sync parallel fan-out con barrier (`sync.WaitGroup`). Send dispara N goroutines, espera a todas, joina errores con `errors.Join`. Sin fail-fast — todos los handlers corren. |
| Topic | `topic.go` | `topic[T]` | `Channel[T]` + `lifecycle.Component` | Async buffered fan-out con **cola y worker dedicado por subscriber** (per-sub model). Cada Subscribe asigna su propio inbox + goroutine; un handler lento solo bloquea su propia cola. Send hace fan-out sequential a TODOS los inboxes via `sendWithPolicy` y agrega errores per-sub con `errors.Join`. Implementa lifecycle.Component; callers wirean via type assertion `ch.(lifecycle.Component)` + `lifecycle.Build`. Subscribe pre-Start difiere el worker hasta Start; post-Start spawn inmediato. |
| Queue | `queue.go` | `queue[T]` | `Channel[T]` + `lifecycle.Component` | Async point-to-point distribution: 1 msg → 1 subscriber via round-robin. Worker pool configurable con `WithWorkerCount(n)`. Caller fire-and-forget. Errores via hook. |
| Null | `null.go` | `null[T]` | `Channel[T]` | Sink `/dev/null` — Send descarta el mensaje y dispara el `ErrorHandler` hook con `ErrDropped`. Subscribe acepta handlers para shape compatibility pero nunca los invoca. Sin estado, sin goroutines, sin buffering, sin lifecycle. Para test doubles o wiring "flujo deshabilitado". |

**Tipos públicos del paquete:**
- `Channel[T]` — interface (Send/Subscribe), en `types.go`.
- `Handler[T] func(ctx, Message[T]) error` — signature del subscriber, en `types.go`.
- `Cancel func()` — handle idempotente retornado por Subscribe, en `types.go`.
- `ErrorHandler func(ctx, msg any, err error)` — hook de observabilidad para impls async/sink, en `types.go`.
- `DeadLetter[T]` envelope (en `message.go` junto a `Message[T]`) + **`WithDLQChannel[T any]` Option** (channel-wide): Topic/Queue publican automáticamente un `DeadLetter[T]` a un `Channel[DeadLetter[T]]` cuando un handler falla. Paralelo a `WithErrorHandler` (observability vs reprocess queue, complementarios). Type parameter T se valida en el constructor vía `extractDLQ` + cassert. Publish es best-effort (errores del DLQ Send se ignoran).
- `Message[T]` — envelope con `Payload T` + `Headers`, en `message.go`.
- `Headers` — 13 campos curados desde Spring Integration (MessageID, CorrelationID, CausationID, ReplyTo, Type, Priority, ContentType, ExpirationTime, SequenceNumber, SequenceSize, Timestamp, Source, Custom). Detalle field-by-field en la memoria [[reference-message-headers]].
- `OverflowPolicy` enum con 4 valores: `OverflowBlock`, `OverflowDropNewest`, `OverflowDropOldest`, `OverflowReject` (default).
- `Options` + Option pattern: `WithBufferSize`, `WithDrainTimeout`, `WithWorkerCount`, `WithErrorHandler`, `WithOverflowPolicy`, `WithDLQChannel[T]` (generic; channel-wide DLQ for Topic/Queue dispatchers).
- `DefaultErrorHandler` / `SilentErrorHandler` — defaults para configurar `WithErrorHandler`, en `functions.go`.
- `StepStatus` enum + `StepResult` + `ChainError` — trace de PipelineChannel, en `errors.go`.
- `Error` struct con sentinels: `ErrSendFailed`, `ErrSubscribeFailed`, `ErrClosed`, `ErrHandlerNil`, `ErrContextNil`, `ErrTimeout`, `ErrDrainTimeout`, `ErrHandlerPanic`, `ErrChainFailed`, `ErrNoSubscribers`, `ErrDropped`, `ErrOverflow`, `ErrBufferFull`.

**Constructores:**
- `NewPipelineChannel[T]() Channel[T]`
- `NewBroadcastChannel[T]() Channel[T]`
- `NewTopicChannel[T](name, opts...) Channel[T]`
- `NewQueueChannel[T](name, opts...) Channel[T]`
- `NewNullChannel[T](opts...) Channel[T]`
- `NewMessage[T](payload, cuids.UID) Message[T]` — constructor de envelope con MessageID y CorrelationID auto-populados.

**Patrón "Channel + lifecycle.Component" via assertion.** `topic` y `queue` implementan `lifecycle.Component` además de `Channel[T]`. El constructor retorna `Channel[T]` (interfaz minimal); callers que quieren lifecycle hacen `ch.(lifecycle.Component)` y wirean con `lifecycle.Build`. Esto mantiene la API de Channel pura y la composición lifecycle opcional. Pattern reusable cuando una primitiva implementa interfaces ortogonales.

**Context propagation en async channels.** `topic` y `queue` no propagan la cancelación del `ctx` del publisher al handler — el dispatch async debe sobrevivir al request del publisher. El handler recibe un ctx fusionado (`mergeContexts` en `context.go`) cuyo Done/Deadline/Err siguen al ctx del worker pero cuyos Value lookups caen primero en el ctx del publisher (propaga trace span, correlation id, slogctx attrs). Detalle en el package doc de `types.go`.

**Per-subscriber model (Topic, YA-0167).** `TopicChannel` aísla a sus subscribers: cada `Subscribe` allocates su propio `chan envelope[T]` (tamaño `WithBufferSize`) y spawna su propio worker goroutine. `Send` fan-outs el mensaje a TODOS los inboxes secuencialmente usando `sendWithPolicy` per-inbox, y agrega errores per-sub con `errors.Join`. Resultado: un handler lento no bloquea a los rápidos (cada uno avanza a su propio ritmo); `OverflowPolicy` y `WithBufferSize` aplican por-subscriber. Subscribe antes de Start difiere el worker hasta Start; post-Start lo spawna inmediato. Cancel cierra `sub.done` (worker exit fire-and-forget); messages in-flight a un sub cancelado se pierden naturalmente. Stop cierra todos los inboxes + cancela workerCtx → workers exit → `awaitDrain` cierra `done`. Sentinel `workerWG.Add(1)` en Start mantiene el WG > 0 para evitar la race de Wait-vs-Add-de-cero con post-Start Subscribes.

**Overflow policy (Topic + Queue).** `OverflowPolicy` selecciona el comportamiento de `Send` cuando el buffer está full:
- `OverflowReject` (**default**) — `Send` retorna `ErrSend(ErrBufferFull)` inmediatamente. Sin block, sin drop, sin hook. Caller decide retry/shed/fallback.
- `OverflowBlock` — `Send` bloquea hasta que abra un slot o ctx expire. Sin pérdida. Comportamiento histórico (pre-YA-0180).
- `OverflowDropNewest` — `Send` retorna `nil`, dropea el mensaje nuevo, fires hook con `ErrOverflow|ErrDropped`.
- `OverflowDropOldest` — `Send` retorna `nil`, evicts head del buffer (fires hook con el evicted msg), enqueue el nuevo. Single retry best-effort si race con otros producers.

**Breaking change YA-0180**: el default cambió de `OverflowBlock` a `OverflowReject`. Código que asuma que `Send` bloquea hasta drain debe pasar `WithOverflowPolicy(OverflowBlock)` explícitamente. La motivación es forzar al operador a tomar una decisión consciente sobre saturación en producción en lugar de heredar bloqueo silencioso. Dispatch implementado en `sendWithPolicy` (`internals.go`); helper compartido entre `topic.Send` y `queue.Send`.

**Estructura de archivos del root package:**
- `types.go` — `Channel[T]` interface + `Handler`/`Cancel`/`ErrorHandler` types + compliance vars + package doc.
- `functions.go` — funciones libres públicas: `DefaultErrorHandler`, `SilentErrorHandler`.
- `internals.go` — helpers libres privados compartidos: `snapshotHandlers`, `invokeHandler`, `invokeStep`, `generateID`, `sendWithPolicy`, `extractDLQ`, `publishDeadLetter`.
- `message.go` — `Message[T]` + `Headers` + `NewMessage` + `DeadLetter[T]` envelope.
- `context.go` — `envelope[T]` + `mergedContext` + `mergeContexts` (concern de propagación async).
- `errors.go` — `Error` struct, sentinels, `ErrXxx` factories, `ChainError` + `StepResult`.
- `options.go` — `Options` + `Option` + `WithXxx`.
- `pipeline.go` + tests.
- `broadcast.go` + tests.
- `topic.go` + tests — incluye `subscriber[T]` struct privado (inbox + done + handler) por R1 (concern del archivo de impl).
- `queue.go` + tests.
- `null.go` + tests.
- `examples/` — submódulo propio con `go.mod`, demos runnable de los 5 channels + features cross-cutting.

### Capa 2: EIP Patterns (sub-paquetes)

Cada patrón vive en su propio sub-paquete bajo `modules/messaging/`, importa el root para usar `Channel[T]`/`Message[T]`/etc. Comparten el mismo `go.mod` que el root. Classification: **Shape B**, struct privado `<name>[T]`, constructor devuelve `lifecycle.Component`. Standard propio en `modules/messaging/CODING_STANDARDS.md`.

| Sub-paquete | Struct | Roles | Qué hace |
|---|---|---|---|
| `bridge/` | `bridge[T]` | `lifecycle.Component` | One-to-one channel forwarder: subscribe a `src`, reenvía cada `Message[T]` a `dst` sin alteración. Patrón identity transform — valor estructural (sync↔async decoupling, etiqueta nombrada en wiring graph, hook point de observabilidad). |
| `filter/` | `filter[T]` | `lifecycle.Component` | Message Filter: subscribe a `src`, reenvía a `dst` solo cuando `PredicateFn` retorna true. Dos hooks separados — `WithErrorHandler` (fallos reales: predicate error/panic, forward fail) y `WithDropHandler` (drops intencionales, silent default). |
| `router/` | `router[T]` | `lifecycle.Component` | Content-Based Router (key → `Channel[T]`): subscribe a `src`, evalúa `RouteFn(msg) → key`, busca destino en `routes[key]` y reenvía. `WithDefaultChannel` opcional para política NoRoute; sin él, NoRoute drops + reporta vía `ErrorHandler`. |
| `idempotent/` | `idempotent[T]` | `lifecycle.Component` | Idempotent Receiver: subscribe a `src`, extrae dedup key via `KeyFn[T]` (default: `Headers.MessageID`), consulta `store.MetadataStore.Has`, reenvía a `dst` solo si la key no fue vista dentro del TTL. Duplicates y keyless dropean via `WithDropHandler` (con `DropReason` — `DropReasonDuplicate` / `DropReasonNoKey`). Fail-closed en `Has` (no forward); fail-open en `Add` (sí forward, log el error). |
| `claimcheck/` | `claimCheckIn[T]` + `claimCheckOut[T]` | `lifecycle.Component` (ambos) | Claim Check (par In + Out): `In` subscribe a `src` (heavy `Message[T]`), guarda original en `store.MessageStore[T]` bajo key generada via `KeyGenFn` (default crypto/rand 128-bit hex), reenvía `Message[ClaimCheckReference]{Key}` a `dst` (preservando `Headers.CorrelationID` del original). `Out` subscribe a `src` (referencias), retrieve original del store, reenvía a `dst` (`Message[T]`), opcionalmente borra del store via `WithDeleteAfterRetrieve` (default true). Fail-closed en Put/Get; fail-open en Delete. |

**Constructores:**
- `NewBridge[T](name, src, dst, opts...) lifecycle.Component`
- `NewFilter[T](name, src, dst, predicate, opts...) lifecycle.Component`
- `NewRouter[T](name, src, decide, routes, opts...) lifecycle.Component`
- `NewIdempotent[T](name, src, dst, metaStore, opts...) lifecycle.Component`
- `NewClaimCheckIn[T](name, src, dst, msgStore, opts...) lifecycle.Component`
- `NewClaimCheckOut[T](name, src, dst, msgStore, opts...) lifecycle.Component`

**Override de la regla universal — errors no propagan al source channel.** Cada pattern subscribe a `src` con un handler que **siempre retorna nil**. Fallos de routing/filtering/forwarding NO son fallos del source channel; el caller del source no debe verlos. Los errores del pattern fluyen vía el `WithErrorHandler` propio (default: `messaging.DefaultErrorHandler` que loguea via `common/log`; opt-out con `messaging.SilentErrorHandler`). Documentado en `modules/messaging/CODING_STANDARDS.md`.

**Lifecycle.** Los patterns implementan `lifecycle.Component` worker-style: Start registra subscription en source channel, Stop la cancela y cierra Done. Ningún pattern spawna goroutines propias — la concurrencia de dispatch se hereda del source channel.

**Estructura de archivos (por sub-paquete):**
- `types.go` — package doc + `Fn` types públicos (`RouteFn`, `PredicateFn`, `DropHandler`, `ErrXxxFn`) + compliance vars.
- `errors.go` — `XxxType` constant, sentinels, `Error` struct, `ErrXxx` factory.
- `options.go` — `Options` + `Option[T]`/`Option` + `WithXxx`.
- `<name>.go` — struct privado `<name>[T]` + constructor `New<Name>` + métodos lifecycle.

**Patrones futuros** (`transformer/`, `splitter/`, `aggregator/`, `scattergather/`, `delayer/`, `wiretap/`, `recipientlist/`, `enricher/`, `headerfilter/`, `history/`, `controlbus/`, `gateway/`, `resequencer/`, `barrier/`) se agregan uno a uno cuando un consumer real los pida. No pre-crear sub-packages vacíos.

### Capa 3: Support primitives (sub-paquetes)

Storage / utility contracts consumidos por los patterns de Capa 2 — no son endpoints ni transforms, son primitivas de soporte. Ship interface + impl canónica in-memory bajo el mismo `go.mod` (paralelo a `core/common/cache/`); backends pesados (Redis, Postgres, S3, …) viven en `extension/messaging/store/<backend>/` con `go.mod` propio para aislar MVS.

| Sub-paquete | Interfaces | Impls in-memory | Consumers EIP | Qué hace |
|---|---|---|---|---|
| `store/` | `MessageStore[T]` (Put/Get/Delete) + `MetadataStore` (Has/Add) | `inMemoryMessageStore[T]` (passive, mutex+map) + `inMemoryMetadataStore` (lifecycle.Component con sweeper goroutine) | Claim Check (YA-0229) consume `MessageStore[T]`; Idempotent Receiver (YA-0228) consume `MetadataStore` | Persiste envelopes completos por key opaca (`MessageStore[T]`) y presencia booleana con TTL (`MetadataStore`). Storage layer reusable por múltiples EIP patterns. |

**Constructores:**
- `NewInMemoryMessageStore[T]() MessageStore[T]` — passive, sin lifecycle.
- `NewInMemoryMetadataStore(name, opts...) MetadataStore` — implementa `lifecycle.Component` por el sweeper; type-assert para wirear vía `lifecycle.Build`.

**Sweeper goroutine (metadata only).** `inMemoryMetadataStore` arranca un único goroutine en `Start` que evicta entries expirados cada `WithSweepInterval` (default 1m). `Has` hace su propia freshness check on-call — un entry expirado pero aún no swept retorna false correctamente. El sweeper sólo libera el slot del map. Stop cierra `stopCh`, espera al goroutine (bounded por `ctx`) y cierra `Done`.

**Errores.** `ErrStore(causes...)` factory + sentinels `ErrStoreFailed`/`ErrStoreNotFound`/`ErrStoreClosed`/`ErrInvalidTTL`. `MessageStore.Get` retorna `ErrNotFound()` (matchable via `errors.Is(err, ErrStoreNotFound)`). `Add` con TTL ≤ 0 retorna `ErrStore(ErrInvalidTTL)`. Operaciones post-Stop sobre `MetadataStore` retornan `ErrStore(ErrStoreClosed)` para distinguir "store muerto" de "key no presente".

**Estructura de archivos:**
- `types.go` — package doc + interfaces (`MessageStore[T]`, `MetadataStore`) + compliance vars + `Fn` types.
- `errors.go` — `StoreType` constant, sentinels, `Error` struct, `ErrStore`/`ErrNotFound` factories.
- `options.go` — `Options` + `Option` + `WithSweepInterval` (consumido sólo por `inMemoryMetadataStore`).
- `messagestore.go` — `inMemoryMessageStore[T]` + `NewInMemoryMessageStore`.
- `metadatastore.go` — `inMemoryMetadataStore` + `NewInMemoryMetadataStore` + sweeper.

## Módulo `modules/core/security/authn/`

Top-level module que aloja el contrato de autenticación + la impl canónica `tokenAuthenticator` (in-module porque sólo depende de `crypto/tokens`, otro módulo del workspace). Transport adapters viven en sus propios go-modules bajo `modules/extension/security/authn/` para que `google.golang.org/grpc` no se filtre vía MVS a consumers que sólo necesitan el contrato. Classification: **Shape B con package único; transports split a módulos hermanos**.

**Tipos / símbolos públicos:**
- `Principal` — struct con `ID`/`Name`/`Roles`/`Attributes`.
- `Authenticator` — interface (`Validate(ctx, token) (*Principal, error)`).
- `Error` + `ErrAuthentication(causes...) error` factory.
- Funciones libres: `WithPrincipal(ctx, *Principal) context.Context`, `FromContext(ctx) (*Principal, bool)`, `NewTokenAuthenticator(method, opts...) Authenticator`.
- Options pattern: `Option`, `Options`, `NewOptions`, `WithSubjectClaim`, `WithNameClaim`, `WithRolesClaim`.
- Fn aliases: `WithPrincipalFn`, `FromContextFn`, `ErrAuthenticationFn` (compliance exhaustiva en `types.go`).
- Sentinels: `ErrAuthenticationFailed`, `ErrTokenEmpty`, `ErrTokenInvalid`, `ErrAuthenticatorNil`, `ErrHeaderMissing`, `ErrHeaderMalformed`, `ErrPrincipalNil`, `ErrMethodNil`, `ErrSubjectClaimMissing`.

**Backend canónico.** `tokenAuthenticator` (struct privado) delega verificación a `*crypto/tokens.Method`. Funciona con los 15 algoritmos soportados por `crypto/tokens` (familia JWT — HS/RS/PS/ES/EdDSA — y familia opaque AEAD — OPAQUE_AES_GCM, OPAQUE_XCHACHA20_POLY1305) porque el dispatch lo hace `Method.Validate`. Mapea claims del Payload a `*Principal` con claves configurables.

**Contrato de fallo.** Todas las impls de `Authenticator` envuelven errores vía `ErrAuthentication(causes...)`. Esto garantiza `errors.Is(err, ErrAuthenticationFailed) == true` para cualquier fallo, permitiendo a los transport adapters traducir uniformemente a 401 / `codes.Unauthenticated` sin inspeccionar errores concretos.

**Aislamiento de dependencias.** Un consumer del contrato no arrastra `google.golang.org/grpc` ni `net/http` server (más allá de stdlib) — esas viven en go-modules separados. Por MVS, cualquier `require` en un go.mod consumido se materializa en el `go.sum` del consumer aunque no compile el código importador — la separación en go-modules es la única que evita esto.

**Sin lifecycle.** El módulo no aloja `lifecycle.Component`. Cada autenticador es un validador stateless. Si una impl futura necesitara caching o background refresh, debe ir a un top-level `modules/managed/<name>/` y exponer `Authenticator` para wirearse contra este contrato.

**Layout plano.** No hay subpaquetes anidados. Futuros backends sin dep externa (apikey, chain, mock, ...) viven en el root junto a `tokenAuthenticator`. Backends con dep externa van a `modules/extension/security/authn/<x>/`. Sub-dominios nuevos (ej. sessions) son **hermanos top-level**: `modules/core/security/sessions/`, no `modules/core/security/authn/sessions/`.

## Módulo `modules/extension/security/authn/http/`

Transport adapter: server-side `net/http` Bearer middleware sobre el contrato `security/authn`. Módulo independiente para mantener `google.golang.org/grpc` fuera del closure de cualquier consumer HTTP-only (y viceversa).

| Paquete | Shape | Externos | Qué hace |
|---|---|---|---|
| `http` | Shape B | `net/http` (stdlib), `security/authn` | `NewMiddleware(authenticator, opts...) Middleware` — `func(http.Handler) http.Handler` que extrae `Authorization: Bearer <token>` (header + scheme configurables), valida vía `Authenticator`, inyecta `*Principal` en `r.Context()`. Fallos → 401 vía `ErrorHandler` configurable (`WithErrorHandler`). |

**Options públicas:** `WithHeaderName(string)`, `WithScheme(string)`, `WithErrorHandler(func(http.ResponseWriter, *http.Request, error))`.

## Módulo `modules/extension/security/authn/grpc/`

Transport adapter: gRPC unary + stream server interceptors sobre el contrato `security/authn`. Módulo independiente para que `google.golang.org/grpc` no se filtre vía MVS a consumers que sólo necesitan el contrato o HTTP.

| Paquete | Shape | Externos | Qué hace |
|---|---|---|---|
| `grpc` | Shape B | `google.golang.org/grpc`, `security/authn` | `NewUnaryInterceptor(authenticator, opts...)` y `NewStreamInterceptor(authenticator, opts...)` — leen el token del metadata key `authorization` (configurable), validan, inyectan `*Principal` en `ctx`. Fallos → `status.Error(codes.Unauthenticated, ...)`. Stream wraps el server stream con `Context()` override para propagar el ctx con el Principal sin reasignar. |

**Options públicas:** `WithMetadataKey(string)`, `WithScheme(string)`.

