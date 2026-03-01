# Brainstorm

Ideas, trabajo futuro y extensiones potenciales.
Nada aqui es un compromiso — son direcciones posibles cuando haya un caso de uso concreto.

---

## crypto

### X.509 / PKI

Paquete dedicado para certificados y claves en formatos estandar.

- Parsear y generar certificados X.509 (`ParseCertificate`, `CreateCertificate`, DER/PEM)
- Parsear/generar claves privadas: RSA (PKCS#1, PKCS#8), ECDSA (PKCS#8, SEC1), Ed25519 (PKCS#8)
- CSRs — PKCS#10 (`ParseCertificateRequest`, `CreateCertificateRequest`)
- Extensiones X.509: BasicConstraints, KeyUsage, ExtendedKeyUsage, SAN, AKI/SKI, CRL, OCSP
- Algoritmos de firma: RSA-PSS, ECDSA (P-256, P-384, P-521), Ed25519

Esto seria un modulo o paquete grande por si solo, no un subpaquete de `crypto/`.

### Cifrado hibrido (ECIES / HPKE)

Combinaciones de key agreement + KDF + AEAD:

- `ECDH_P256 + HKDF_SHA256 + AES-GCM`
- `ECDH_P521 + HKDF_SHA512 + AES-GCM`
- `X25519 + HKDF_SHA256 + ChaCha20-Poly1305`

Requiere implementar ECDH, X25519 y HKDF como building blocks.
Solo tiene sentido con un caso de uso real (E2E encryption, key encapsulation).

### KDF

- HKDF sobre SHA-256 / SHA-512 (key derivation)

### Referencia de nombres

Cada algoritmo implementado tiene equivalencias en distintos estandares:

| Interno | JOSE/JWT | TLS | OpenSSL |
|---------|----------|-----|---------|
| `HMAC_with_SHA256` | HS256 | — | — |
| `HMAC_with_SHA512` | HS512 | — | — |
| `ECDSA_with_SHA256_over_P256` | ES256 | ecdsa-with-SHA256 | — |
| `ECDSA_with_SHA512_over_P521` | ES512 | ecdsa-with-SHA512 | — |
| `RSASSA_PSS_using_SHA256` | PS256 | sha256WithRSAandMGF1 | rsa_padding_mode:pss -sha256 |
| `RSASSA_PSS_using_SHA512` | PS512 | sha512WithRSAandMGF1 | rsa_padding_mode:pss -sha512 |
| `Ed25519` | EdDSA (crv:Ed25519) | Ed25519 | -ed25519 |
| `AES_128_GCM` | A128GCM | TLS_AES_128_GCM_SHA256 | aes-128-gcm |
| `AES_256_GCM` | A256GCM | TLS_AES_256_GCM_SHA384 | aes-256-gcm |
| `ChaCha20-Poly1305` | — | TLS_CHACHA20_POLY1305_SHA256 | chacha20-poly1305 |
| `RSA-OAEP-SHA256` | RSA-OAEP-256 | — | rsa_oaep -sha256 |
| `RSA-OAEP-SHA512` | RSA-OAEP-512 | — | rsa_oaep -sha512 |

---

## common

### HTTP client — evolución

Wrapper HTTP basado en `http.Client` con `rate.Limiter`. Fase 0 (hardening, contrato, limiter, body close) completada.
Observabilidad y ergonomía pospuestas para una capa superior.

Pendiente:

- **Resiliencia**: `Retry-After`, backoff exponencial con jitter, `Idempotency-Key`, circuit breaker
- **Casos de uso**: rate por host/endpoint, auth con refresco automatico (bearer/JWT/API key), headers especializados (`X-Request-Id`, `ETag`), descargas con `Range`, subidas multipart
- **Concurrencia**: worker pool con limiter compartido, priorizacion, backpressure
- **Seguridad**: redaccion de secretos en logs, TLS segura en `Transport`, sanitizacion de errores
- **Rendimiento**: tuning de `Transport` (`MaxConnsPerHost`, `IdleConnTimeout`), benchmarks, profiling

Referencia original: inspirado en [grpc-java examples](https://github.com/grpc/grpc-java/tree/master/examples).

### Optional[T]

Generic Optional/Maybe type que wrappea un valor con flag de presencia.
Go idiomatico usa punteros para opcionalidad, pero el pattern puede ser util para APIs
donde nil pointer no es deseable y se quiere un `HasValue()`/`Value()`/`Default()` explicito.

Evaluar si agrega valor sobre `*T` o si es over-engineering.

---

## managed

### Clasificacion por paquetes

Posible organizacion de managed types:

- **workers** — base workers con lifecycle generico
- **servers** — HTTP, gRPC y otros servidores de red
- **consumers** — Kafka, RabbitMQ, NATS y otros message consumers
- **pollers** — polling-based workers
- **resources** — managed resources con lifecycle

---

## telemetry

### Datadog integration

Alternativa a OTel para observabilidad. Dos versiones exploradas:

- **v1**: `gopkg.in/DataDog/dd-trace-go.v1` — tracer, profiler, metrics client. Fatal on failure.
- **v2**: `github.com/DataDog/dd-trace-go/v2` — misma estructura pero retorna errors en vez de fatal.

Solo considerar si hay un caso de uso donde OTel no es suficiente o el stack es Datadog-native.

---

## modulos nuevos (de go-feather-lib)

### `modules/datasource/`

Abstracciones de base de datos migradas de go-feather-lib:

- **gorm** — Connection, Context, TransactionHandler para SQL via GORM
- **mongo** — Connection, Context, TransactionHandler para MongoDB
- **goredis** — Connection, Context, TransactionHandler para Redis
- **gocql** — Connection, Context para Cassandra

Patron comun: Context (url/server/credentials), Connection (open/close), TransactionHandler (callback-based).
Requiere disenar como managed components con lifecycle.

### `modules/messaging/`

Dos capas separadas en go-feather-lib que deberian unificarse en yarumo:

#### Capa 1 — Abstraccion generica (go-feather-lib `integration/messaging/`)

Enterprise Integration Patterns (EIP) con generics:

**Tipos core:**
- `Message[T]` interface — `Headers()` + `Payload()` + `String()`
- `Headers` interface — 15+ campos tipados: Id, MessageType, Timestamp, ContentType, Expired, TTL, Origin/Destination/Reply/ErrorChannel, CorrelationId, etc.
- `ErrorPayload` interface — `Code()`, `Message()`, `Errors()`
- `ErrorMessage[T]` interface — wraps error + original message

**Channel interfaces:**
- `SenderChannel[T]` — `Send(ctx, timeout, message)`, `Name()`
- `ReceiverChannel[T]` — `Receive(ctx, timeout)`, `Name()`
- `MessageChannel[T]` — combines Sender + Receiver
- Handler function types: `SenderHandler[T]`, `ReceiverHandler[T]`

**Implementaciones:**
- `QueueChannel[T]` — buffered `chan Message[T]` con tracking de expiry
- `PassThroughChannel[T]` — wraps sender + receiver separados
- `FunctionAdapterSenderChannel[T]` / `FunctionAdapterReceiverChannel[T]` — wraps handler functions

**Decorators (stackeables):**
- `TimeoutSenderChannel[T]` / `TimeoutReceiverChannel[T]` — context timeout con goroutine
- `LoggedSenderChannel[T]` / `LoggedReceiverChannel[T]` — trace logging
- `HeadersValidatorSenderChannel[T]` / `HeadersValidatorReceiverChannel[T]` — validacion de headers antes de send/receive

**Composicion:** `BaseMessageChannel[T]()` stackea: queue -> validator -> timeout -> logged.

**Nota:** En go-feather-lib esta capa y los drivers de rabbitmq son independientes (no se conectan).
En yarumo deberian unificarse: los drivers implementan los channel interfaces.

#### Capa 2 — Drivers de message brokers (go-feather-lib `messaging/`)

**rabbitmq/amqp** (amqp091-go):
- `Context` (url, server, vhost) + `Connection` con retry (5 attempts, 2s delay)
- `Consumer` — queue declaration, delivery processing, closing handler, flags (autoAck, durable, exclusive, etc.)
- `Producer` — exchange publish con mandatory/immediate flags
- `ConsumerServer` — spawns goroutine por consumer, signal handling
- `Dialer()` / `DialerTLS()` — connection factories
- Options pattern completo con chains

**rabbitmq/streams** (rabbitmq-stream-go-client):
- Mismo patron que AMQP pero usa `stream.Environment` en vez de `amqp.Connection`
- Consumer con stream declaration, offset tracking
- Producer con stream publish
- `ConsumerServer` identico al de AMQP

**kafka** — placeholder (no implementado en go-feather-lib)
**nats** — placeholder (no implementado en go-feather-lib)

#### Diseno sugerido para yarumo

```
modules/messaging/
  ├── types.go          — Message[T], Headers, Channel interfaces (de integration/)
  ├── channels/         — QueueChannel, decorators (de integration/)
  ├── rabbitmq/
  │   ├── amqp/         — driver que implementa Channel interfaces
  │   └── streams/      — driver que implementa Channel interfaces
  ├── kafka/            — futuro
  └── nats/             — futuro
```

Requiere disenar como managed consumers/producers con lifecycle.

### `modules/auth/`

Sistema de autenticacion/autorizacion de go-feather-lib:

- **AuthenticationService** — Authenticate (credentials -> Principal), Validate (token -> Principal)
- **AuthorizationFilter** — HTTP middleware que extrae token, autoriza, inyecta Principal en context
- **PrincipalManager** — CRUD de principals (base + gorm implementations)
- **DelegatingPasswordEncoder** — routea por prefijo: `{bcrypt}hash`, `{argon2}hash`, `{scrypt}hash`

Depende de: passwords, tokens, datasource (para persistence).
El DelegatingPasswordEncoder podria ir directo en `common/crypto/passwords` como quick win.

### `common/cache/`

go-feather-lib tenia un directorio `cache/` vacio y un import comentado de `gomemcache`.
Nunca se implemento. Si se necesita, disenar desde cero:

Posibles direcciones:
- **In-memory** — LRU/LFU con TTL (stdlib `sync.Map` o `hashicorp/golang-lru`)
- **Distributed** — abstraccion sobre Redis (ya existe `datasource/goredis`) o Memcached
- **Interface minima:** `Get(key)`, `Set(key, value, ttl)`, `Delete(key)`, `Has(key)`

Evaluar si es un paquete de `common/` o un managed component.

### `common/health/`

Health checking de go-feather-lib:

- `Health` interface con `GetStatus()`, `RegisterChecker()`, `ServeHTTP()`, `Shutdown()`
- `Checker` interface para verificar servicios externos (DB, cache, etc.)
- Memory stats, uptime, goroutine count
- HTTP handler para health endpoints

Util para produccion. Podria ser `common/health/` o un managed component.

### `common/stats/`

Funciones estadisticas de go-feather-lib:

- Aritmeticas: Add, Sub, Multi, Div, Mod, Pow, Sqrt, Cbrt, Log, Abs, Ceil, Floor, Round
- Agregaciones: Sum, Count, Average
- Medidas de tendencia central: Mean (arithmetic, geometric, harmonic, quadratic), Median, Mode

Ya existe `maths/sets` y `maths/logic`. Podria ser `maths/stats/` si hay caso de uso.

### `common/validation/`

Validacion de structs/fields de go-feather-lib:

- `ValidateStructIsRequired()`, `ValidateStructMustBeUndefined()`
- `ValidateFieldIsRequired[T]()`, `ValidateFieldMustBeUndefined[T]()`
- Error factories para cada caso

Go ya tiene `go-playground/validator`. Evaluar si agrega valor sobre eso.

---

## ODM — Operational Decision Manager

### Vision

Sistema de decisiones basado en reglas sobre logica proposicional.
Se construye en capas sobre `modules/maths/logic` y `modules/inference`.

### Arquitectura por capas

```
app/odm-api          ← decision service (HTTP/gRPC, auth, persistencia)
app/odm-console      ← UI para rule authoring, testing, audit

modules/inference     ← engine puro (forward chaining, explain)
    │
    ▼
modules/maths/logic   ← matematica pura (SAT, parsing, entailment)
    │
    ▼
modules/common
```

### Que va en cada capa

| Capa | Responsabilidad | Reutilizable? |
|------|----------------|---------------|
| `maths/logic` | SAT, parsing, simplificacion, entailment | Alto |
| `inference` | Engine (forward chaining), rules, explain, ruleset serialization (JSON/YAML) | Alto |
| Binding struct↔Fact | Convertir structs/JSON de negocio a `logic.Fact` | Bajo — cada dominio es distinto |
| Decision service | API HTTP/gRPC, persistencia de rulesets, audit trail | Nulo — es una app |
| Rule authoring UI | Editor de reglas, testing, simulacion | Nulo — es una app |

### Por que no un sdk/decisions

El binding entre el mundo real (structs, JSON, campos de negocio) y las variables
proposicionales es inherentemente especifico de cada dominio. Lo que para un proyecto
es `cliente.edad >= 18`, para otro es `sensor.temperatura > 100`.

Las mecanicas genericas reutilizables (cargar rulesets, versionar, ejecutar contra
`map[string]any`, serializar traces) son tan delgadas que caben dentro de `inference`
mismo. Un SDK separado no se justifica — terminaria siendo un framework tipo Drools/IBM ODM
que intenta resolver el binding para todos, y esos sistemas son monstruos por esa razon.

### Que necesitaria cada app

1. **Modelo de dominio (rule authoring)** — DSL o schema para definir reglas en terminos
   de negocio. Ejemplo: `"si cliente.edad >= 18 AND cliente.ingresos > 5000 THEN aprobar_credito"`.
   Binding entre variables proposicionales y campos de structs/JSON.

2. **Repositorio de reglas** — almacenamiento persistente (DB, archivos, API), versionado
   de rulesets, agrupacion ("ruleset de credito v2.3").

3. **Decision service (API)** — endpoint que recibe hechos (JSON), ejecuta un ruleset,
   retorna conclusiones. Stateless, idempotente.

4. **Integracion con datos reales** — adaptador que convierte struct/JSON de entrada en
   `logic.Fact`, adaptador que convierte conclusiones del engine en respuestas tipadas.

5. **Audit/compliance** — persistencia del trace de explain: que reglas dispararon, en que
   orden, con que datos. Quien cambio que regla, cuando.

6. **Testing de reglas** — "dado estos hechos, espero estas conclusiones" — tests de
   regresion sobre rulesets. Simulacion: "que pasaria si cambio esta regla?".

### Rol de la IA en el ODM

La IA es conveniencia, no arquitectura. El ODM funciona igual sin ella — solo que el
usuario escribe las reglas a mano y lee los traces crudos. La IA es la interfaz humana
del ODM, no el motor. El valor del motor es que no es probabilistico.

**Donde entra la IA:**

| Tarea | Que hace | IA local funciona? |
|-------|---------|-------------------|
| NL → facts (campos conocidos) | Extrae datos estructurados de texto libre | Si — 7-8B basta |
| Trace → explicacion NL | Reformatea trace tecnico a lenguaje humano | Si — 7-8B basta |
| NL → reglas (authoring) | Genera `A & B => C` desde lenguaje natural | Depende de la complejidad |
| Validacion de rulesets | Detecta contradicciones, casos no cubiertos | No necesita IA — `maths/logic` lo hace (SAT, entailment) |

**Lo que la IA NO hace:** no ejecuta reglas, no decide, no reemplaza auditabilidad.
"Por que se rechazo este credito" tiene que ser una cadena de reglas trazable, no
"el modelo dijo que no".

**Seleccion de modelo — criterios:**

1. Structured output confiable (JSON/formato parseble, no prosa)
2. Latencia vs costo: authoring es offline (puede ser lento), extraction puede ser online
3. Modelo chico vs grande: extraccion y formateo → chico; generacion de reglas ambiguas → grande

**Candidatos locales (para extraccion y formateo):**

- Qwen 2.5 (7B/14B) — muy bueno en structured output
- Llama 3.1 (8B) — el mas probado, comunidad enorme
- Mistral Small (22B) — buen balance tamaño/capacidad
- Phi-3 (3.8B/14B) — sorprendentemente capaz para su tamaño

**Candidatos cloud (para generacion de reglas complejas):**

- Haiku/Sonnet — suficiente para extraccion y formateo, rapido y barato
- Opus — solo si hay ambiguedad compleja en el input

**Estrategia recomendada:**

1. Empezar con Sonnet via API para todo
2. Definir 20 ejemplos input/output por tarea
3. Medir tasa de error
4. Bajar a Haiku o local lo que funcione
5. No optimizar antes de medir

### Conclusion

`inference` es el motor — puro, sin opiniones de infraestructura. El ODM es producto
que usa el motor. No hay modulo intermedio que justifique existir como libreria reutilizable.
La frontera clara es: `inference` resuelve "que se puede concluir", cada app resuelve
"como conecto eso con mi negocio".

---

## maths — debates pendientes

### Regresion: donde vive?

Regresion lineal (simple y multiple) aparece en el plan de estudio como puente entre
estadistica y ML. Matematicamente es algebra lineal pura: beta = (X^T X)^(-1) X^T y.
Pero su uso es estadistico (inferencia sobre parametros, R^2, residuales) y tambien
predictivo (modelo base de ML).

**Opciones:**

| Opcion | Ubicacion | Argumento |
|--------|-----------|-----------|
| A | `maths/stats/` | Regresion como herramienta estadistica (mínimos cuadrados, R^2, intervalos) |
| B | `inference/regression/` | Regresion como motor predictivo, al mismo nivel que bayesian/fuzzy/classical |
| C | Ambos | stats/ tiene las primitivas (least squares), inference/ tiene el motor (fit, predict, explain) |

**A favor de C:** sigue el patron establecido: maths/ = primitivas, inference/ = motores.
Pero puede ser over-engineering si regresion es simple.

**Pendiente:** decidir cuando haya un caso de uso concreto.

### Geometric Algebra: alternativa a linalg/?

GA (Geometric Algebra / Clifford Algebra) unifica y simplifica conceptos de algebra lineal:
rotaciones sin matrices (rotores), reflexiones como operacion primitiva, determinantes como
producto exterior, subespacios como blades.

**Pregunta:** si se implementa GA, reemplaza a linalg/ o lo complementa?

- GA subsume producto punto, producto cruz, determinantes, rotaciones, proyecciones
- Pero ML/AI habla el lenguaje de matrices (SVD, eigenvalues, backprop). No se puede ignorar linalg/
- GA seria mas util para: graficos, robotica, simulacion fisica, geometria computacional

**Opciones:**

| Opcion | Ubicacion | Argumento |
|--------|-----------|-----------|
| A | `maths/ga/` | Paquete independiente, alternativa elegante para problemas geometricos |
| B | No implementar | Hobby/nicho, no justifica el esfuerzo para los casos de uso actuales |
| C | `maths/linalg/ga/` | Subpaquete de linalg que ofrece la perspectiva GA sobre las mismas operaciones |

**Pendiente:** decidir despues de estudiar GA (Fase 2 Exploracion Matematica).

---

## coding standards — propuestas

Reglas candidatas que difieren del `CODING_STANDARDS.md` actual. Requieren revision y decision.

### Constructores: campos requeridos solo para singletons/beans

**Propuesta:** Los constructores con parametros requeridos (no-opts) deben reservarse para singletons y beans (managed components). Las instancias regulares deben usar solo el patron options.

**Actual:** Criterio #4 permite parametros requeridos con `assert.NotNil` en cualquier constructor, sin distinguir singleton de instancia regular.

**Ejemplo propuesto:**
```go
// Singleton/bean — campos requeridos permitidos
func NewServer(listener net.Listener, handler http.Handler, options ...Option) Server

// Instancia regular — solo options
func NewMethod(name string, options ...Option) *Method   // ← crypto ya sigue esto
func NewClient(options ...Option) Client                 // ← solo opts
```

**Preguntas abiertas:**
- Como definimos formalmente "singleton/bean" vs "instancia regular"?
- Que pasa con constructores como `NewMethod(name, kind, keySize, ...Option)` en crypto? — el `name` es requerido pero no es un singleton.

### Asserts solo en constructores

**Propuesta:** Eliminar `assert.NotNil` del receiver en metodos de struct. Los asserts solo aplican en constructores.

**Actual:** Criterio #4 requiere `assert.NotNil` en el receiver al inicio de cada metodo de struct:
```go
func (m *Method) Hash(data []byte) ([]byte, error) {
    assert.NotNil(m, "method is nil")       // ← se eliminaria
    assert.NotNil(m.hashFn, "hashFn is nil") // ← se eliminaria
    ...
}
```

**Preguntas abiertas:**
- Si el constructor valida todo, un receiver nil solo es posible por uso incorrecto (`var m *Method; m.Hash(...)`). Vale la pena el assert?
- Crypto ademas valida function fields (`m.hashFn`). Estas se setean en el constructor via options — si el constructor garantiza defaults, el assert es redundante.
- Impacto: hay asserts de receiver en todos los paquetes revisados (crypto, http, grpc, rest, cron, etc). Seria un cambio amplio.

---

## tools

### RouteGen — Code generation para rutas Gin

Herramienta de code generation que a partir de metodos anotados con `@route METHOD /path`
genera funciones de definicion de rutas para Gin.

**Input:**
```go
type handlers struct {
    repository Repository
}

// @route POST /events
func (h *handlers) CreateEvent(c *gin.Context) {}

// @route GET /events/:id
func (h *handlers) GetEvent(c *gin.Context) {}
```

**Output** (`zz_routes_gen.go`):
```go
// Code generated by routegen; DO NOT EDIT.
package core

import "github.com/gin-gonic/gin"

func Route_POST_events(h *handlers) (string, string, gin.HandlerFunc) {
    return "POST", "/events", h.CreateEvent
}

func Route_GET_events_id(h *handlers) (string, string, gin.HandlerFunc) {
    return "GET", "/events/:id", h.GetEvent
}
```

**Reglas de entrada:**
- Receiver debe ser el tipo indicado por `--type`
- Firma exacta: `func(*gin.Context)` (1 param, 0 returns)
- Comentario con `@route METHOD PATH` (METHOD in GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS, PATH inicia con `/`)
- Metodos sin `@route` se ignoran
- Duplicados `(METHOD, PATH)` son error

**Convencion de nombres:**
- Formato: `Route_<METHOD>_<sanitized_path>`
- Sanitizacion: quitar `/` inicial, reemplazar `/` `:` `{` `}` por `_`, colapsar `_`, vacio -> `root`, digito inicial -> `p_` prefijo
- Colisiones: sufijo contador (`_2`, `_3`)

**Arquitectura:**
```
cmd/routegen/main.go
internal/routegen/
  reader/       — discover .go files, AST parse, inspect receiver methods, parse @route, validate
  generator/    — namer (sanitize path), model builder, code emitter, go/format
  shared/       — types (RouteModel, RouteDef, MethodNode), errors
```

**Reader pipeline:** discover -> parse AST -> inspect type methods -> parse annotations -> validate duplicates -> RouteModel
**Generator pipeline:** order by token.Pos -> name functions -> emit code -> format -> write file

**CLI:**
```bash
go install github.com/<org>/routegen/cmd/routegen@latest

# En el paquete:
//go:generate routegen --type handlers --dir . --out zz_routes_gen.go
```

**Flags:** `--dir` (default `.`), `--type` (requerido), `--out` (default `zz_routes_gen.go`), `--func-prefix` (default `Route_`)

**Fuera de scope por ahora:**
- `RegisterRoutes(r gin.IRoutes, h *handlers)` — registrador opt-in
- Grupos (`/v1`)
- Middlewares por anotacion
- Multiples tipos receiver

---

## Aplicaciones y Sinergias — yarumo + ODM + aluna

Lluvia de ideas combinando las capacidades de `modules/maths/` (logic, probability, fuzzy),
`modules/inference/` (classical, bayesian, fuzzy) y el proyecto aluna (SaaS agentico para
PyMEs LATAM) con la vision ODM documentada arriba.

**Observacion clave:** todas las ideas son realizables con el estado actual de maths/ e
inference/ (100% coverage, 3 engines operativos). Las extensiones propuestas en el roadmap
(constraint/, predicate/, statistical/) abririan dimensiones adicionales pero no son
prerrequisito para ninguna idea listada aqui.

### Categoria 1: Aplicativos nuevos — independientes de ODM y aluna

Productos y herramientas construidos con yarumo en dominios no relacionados con ODM ni aluna.

| # | Nombre | Descripcion | Paquetes yarumo | Complejidad | Hoy? |
|---|--------|-------------|-----------------|-------------|------|
| 1.1 | **SmartTune** | Libreria Go para controladores fuzzy PID en IoT/embebidos | `inference/fuzzy` (Mamdani) | Low | Si |
| 1.2 | **StructHealth** | Monitoreo estructural: Bayes para probabilidad de fallo + reglas clasicas para acciones | `inference/bayesian` + `inference/classical` | Medium | Si |
| 1.3 | **LogicLab** | Tutor interactivo de logica proposicional (truth tables, SAT, transformaciones paso a paso) | `maths/logic` (todos los sub-paquetes) | Low | Si |
| 1.4 | **InferencePlayground** | Sandbox multi-paradigma: definir un problema en 3 paradigmas y comparar resultados con trazas | los 3 engines + los 3 maths | Medium | Si |
| 1.5 | **AlertGuard** | Correlacion de alertas infra: Bayes para causa raiz + backward chaining para remediacion | `inference/bayesian` + `inference/classical` | Medium | Si |
| 1.6 | **ConfigValidator** | CLI que valida YAML/JSON configs contra reglas declarativas usando SAT y forward chaining | `maths/logic`, `maths/logic/sat`, `inference/classical` | Low | Si |
| 1.7 | **RiskGrade** | Scorer de riesgo crediticio: fuzzy (scoring gradual) + Bayes (probabilidad de default) | `inference/fuzzy` + `inference/bayesian` | Medium | Si |
| 1.8 | **FraudNet** | Deteccion de anomalias en transacciones: Bayes + fuzzy + reglas clasicas de escalamiento | los 3 engines | High | Si |

**Detalles:**

- **SmartTune** — Un controlador fuzzy clasico: variables de entrada (error, delta-error),
  reglas Mamdani ("si error es grande-positivo Y delta es creciente ENTONCES accion es
  reducir-mucho"), defuzzificacion centroide. El engine fuzzy de yarumo cubre esto completo.
  Valor: libreria plug-and-play para Go en IoT, donde las alternativas son C/C++ o Python.

- **StructHealth** — Red bayesiana modela P(fallo | vibracion, corrosion, carga). Si
  P(fallo) supera umbral, classical engine dispara reglas de accion ("inspeccionar",
  "cerrar", "alertar"). Combina incertidumbre con acciones deterministicas.

- **LogicLab** — App web/CLI educativa. Usa `maths/logic/parser` para leer formulas,
  `maths/logic/sat` para resolver, `maths/logic/entailment` para verificar. Muestra
  transformaciones NNF/CNF/DNF paso a paso. Target: estudiantes de logica/CS.

- **InferencePlayground** — "Rosetta Stone" de inferencia. Define un problema (ej:
  diagnostico medico) y lo resuelve con forward chaining (reglas), Bayes (probabilidades),
  y fuzzy (grados). Compara trazas lado a lado. Herramienta de aprendizaje y evaluacion.

- **AlertGuard** — Correlacionador de alertas para SRE/DevOps. Red bayesiana con
  P(causa_raiz | alerta1, alerta2, ...) + backward chaining para plan de remediacion
  ("si causa es disco-lleno ENTONCES limpiar-logs Y escalar-storage"). Reduce fatiga
  de alertas al agrupar sintomas en causas.

- **ConfigValidator** — `configval validate --rules rules.yaml config.yaml`. Las reglas
  se expresan como proposiciones ("si database.ssl es false Y env es production ENTONCES
  violacion"). SAT verifica consistencia del ruleset. Forward chaining evalua el config.

- **RiskGrade** — Motor de scoring financiero. Fuzzy evalua variables con bordes difusos
  ("ingreso medio-alto", "antiguedad reciente"). Bayes calcula P(default | score, historial).
  Resultado: score numerico + probabilidad + traza explicable.

- **FraudNet** — Pipeline de deteccion: Bayes prioriza transacciones sospechosas,
  fuzzy gradua la severidad, classical engine aplica reglas de escalamiento
  ("si severidad > alta Y monto > umbral ENTONCES bloquear Y notificar"). Los 3 paradigmas
  cooperan en cascada.

### Categoria 2: Aplicativos con sinergia ODM y/o aluna

Ideas que usan el motor de decisiones ODM, benefician la plataforma aluna, o existen en
la interseccion de ambos. Cada idea indica con que se solapa.

| # | Nombre | Descripcion | Overlap | Complejidad | Hoy? |
|---|--------|-------------|---------|-------------|------|
| 2.1 | **DecisionSkill** | Skill de aluna que wrappea un ruleset ODM — decisiones deterministicas auditables como tool del agente | ODM engine dentro de aluna | Medium | Si |
| 2.2 | **SmartValidator** | Vertical aluna: validacion de documentos (facturas, contratos) con LLM extraction + Bayes risk + reglas compliance | ODM rules + aluna agents | High | Si |
| 2.3 | **EvalBot** | Vertical aluna: evaluacion de proveedores con fuzzy multi-criterio + Bayes confiabilidad | aluna vertical Compras | Medium | Si |
| 2.4 | **AuditTrail+** | Middleware aluna-runtime: valida cada accion del agente contra ruleset antes de ejecutar | ODM enforcement en aluna | Low-Medium | Si |
| 2.5 | **WorkflowOptimizer** | Vertical aluna: analisis de eficiencia de procesos con Sugeno fuzzy inference | aluna vertical PMO | Low | Si |
| 2.6 | **ClassifyRoute** | Router inteligente: Bayes + fuzzy para dirigir Jobs al Skill optimo automaticamente | mejora plataforma aluna | Medium | Si |

**Detalles:**

- **DecisionSkill** — El agente aluna tiene Skills (tools). DecisionSkill es un Skill
  generico que carga un ruleset ODM y lo ejecuta. El agente pasa hechos (JSON del contexto),
  recibe conclusiones deterministicas + traza. Ejemplo: "evaluar si este cliente califica
  para descuento" — el LLM extrae datos, DecisionSkill aplica reglas, el resultado es
  auditable y reproducible. El agente NO decide, solo orquesta.

- **SmartValidator** — Pipeline: LLM extrae campos de documento (factura: NIT, monto, fecha,
  items) -> red bayesiana P(fraude | inconsistencias) -> reglas clasicas de compliance
  (formato NIT valido, fecha no futura, items con IVA). Tres capas: extraccion (LLM),
  riesgo (Bayes), cumplimiento (classical). Vertical natural para CierreExpress.

- **EvalBot** — Evaluacion multi-criterio de proveedores. Variables fuzzy: precio (bajo,
  medio, alto), calidad (mala, aceptable, excelente), tiempo-entrega (rapido, normal, lento).
  Reglas fuzzy Mamdani: "si precio es bajo Y calidad es excelente ENTONCES score es muy-alto".
  Bayes agrega P(cumplimiento_historico). Output: ranking con explicacion.

- **AuditTrail+** — Antes de que el agente ejecute una accion (enviar email, modificar
  registro, llamar API), el middleware evalua un ruleset: "puede skill X ejecutar accion Y
  en contexto Z?". Forward chaining con reglas de permiso/restriccion. Traza completa de
  por que se permitio o bloqueo. Seguridad y compliance como reglas, no como codigo.

- **WorkflowOptimizer** — Sugeno inference para evaluar eficiencia de procesos. Variables:
  tiempo-ciclo, tasa-error, utilizacion-recursos. Output: score numerico via Sugeno
  (promedio ponderado). Identifica cuellos de botella con explicaciones fuzzy legibles.

- **ClassifyRoute** — Cuando llega un Job a aluna, el router decide que Skill lo maneja.
  Hoy es estatico. ClassifyRoute usa: Bayes P(skill_optimo | tipo_job, contexto, historial)
  + fuzzy confidence scoring. Si confidence es baja, escala a humano. Mejora la
  orquestacion sin cambiar la arquitectura de Skills.

### Categoria 3: Como ODM se potencia con yarumo y aluna

Formas concretas en que `maths/`, `inference/` y conceptos de aluna hacen al ODM mas
poderoso que un sistema de reglas basico.

| # | Nombre | Fuente | Descripcion | Hoy? |
|---|--------|--------|-------------|------|
| 3.1 | **SAT Validation** | `maths/logic/sat` | Validar consistencia de rulesets con DPLL antes de deploy — detectar contradicciones | Si |
| 3.2 | **Coverage Analysis** | `maths/logic/sat` | Enumerar combinaciones de inputs y encontrar gaps (escenarios sin regla definida) | Si |
| 3.3 | **Rule Simplification** | `maths/logic` | Aplicar 18 reglas de simplificacion a condiciones complejas — legibilidad + comparacion | Si |
| 3.4 | **Dual-Mode Execution** | `inference/classical` | Exponer forward (data-driven) + backward (goal-driven) chaining como 2 modos de query | Si |
| 3.5 | **Bayesian Overlay** | `inference/bayesian` | Asociar confianza probabilistica a reglas empiricas — P(rule_applicable \| data_quality) | Si |
| 3.6 | **Fuzzy Thresholds** | `inference/fuzzy` | Reemplazar umbrales duros (`income > 5000`) con condiciones fuzzy (`income IS high`) | Si |
| 3.7 | **AI Rule Authoring** | aluna patron | Usar patron THINK-ACT-OBSERVE para authoring NL->rules con validacion SAT en el loop | Si |
| 3.8 | **Trace-to-NL** | aluna patron | Convertir trazas tecnicas a explicaciones en espanol via Claude API — auditores y reguladores | Si |

**Detalles:**

- **SAT Validation** — Antes de deployar un ruleset, convertir todas las reglas a formulas
  proposicionales y verificar con DPLL que no hay contradicciones. Ejemplo: si regla A
  concluye "aprobar" y regla B concluye "rechazar" bajo las mismas condiciones, SAT lo
  detecta. Esto es imposible con un motor de reglas puro — requiere la capa matematica.

- **Coverage Analysis** — Dado un ruleset y sus variables de entrada, enumerar todas las
  combinaciones posibles (o un subconjunto via sampling) y verificar que cada combinacion
  tiene al menos una regla que dispara. Identifica "puntos ciegos" donde el sistema no
  tiene respuesta. Usa `maths/logic/sat` para eficiencia.

- **Rule Simplification** — Las 18 reglas de simplificacion de `maths/logic` (absorcion,
  idempotencia, De Morgan, etc.) pueden simplificar condiciones complejas para hacerlas
  legibles. Tambien permite comparar si dos reglas son logicamente equivalentes aunque
  tengan formas distintas (via canonicalizacion).

- **Dual-Mode Execution** — Forward chaining: "dados estos hechos, que puedo concluir?"
  (data-driven, batch). Backward chaining: "es posible concluir X? que falta?"
  (goal-driven, interactivo). El engine classical de yarumo soporta ambos. ODM puede
  exponer ambos modos como endpoints distintos del decision service.

- **Bayesian Overlay** — Las reglas clasicas son binarias (dispara o no). Con un overlay
  bayesiano, cada regla tiene una confianza asociada: P(regla_correcta | calidad_datos).
  Si los datos de entrada son inciertos (ej: OCR con baja confianza), el overlay reduce
  la confianza de las conclusiones proporcionalmente. No reemplaza reglas — las enriquece.

- **Fuzzy Thresholds** — En vez de `income > 5000` (que rechaza a alguien con 4999),
  fuzzy define `income IS high` con transicion gradual. El motor fuzzy calcula grado de
  pertenencia y las reglas operan sobre grados. Output: "aprobado con grado 0.73" en
  vez de "aprobado/rechazado". Mas justo, mas explicable.

- **AI Rule Authoring** — El patron agentico de aluna (THINK-ACT-OBSERVE) se aplica al
  authoring de reglas. THINK: analizar requerimiento NL. ACT: generar regla proposicional.
  OBSERVE: validar con SAT (consistencia) y entailment (no-redundancia). Si falla,
  iterar. El loop garantiza que las reglas generadas por IA son formalmente correctas.

- **Trace-to-NL** — Los traces de `inference/*/explain` son tecnicos (lista de reglas,
  hechos, pasos). Claude API puede convertirlos a explicaciones legibles: "Su solicitud
  fue rechazada porque su ingreso mensual ($3,200) no cumple el minimo requerido ($5,000)
  segun la politica de credito v2.3, regla 4." Auditores y reguladores necesitan esto.

### Categoria 4: Como extender, mejorar y ampliar aluna

Formas en que yarumo y ODM mejoran los verticales actuales de aluna, habilitan verticales
nuevos, o mejoran la plataforma base.

| # | Nombre | Tipo | Descripcion | Paquetes yarumo | Complejidad | Hoy? |
|---|--------|------|-------------|-----------------|-------------|------|
| 4.1 | **Reconciliation Engine** | Mejora CierreExpress | Reglas clasicas para matching deterministico de transacciones, LLM solo para ambiguos — reduce API calls 60-80% | `inference/classical` | Medium | Si |
| 4.2 | **Variance Analysis** | Mejora CierreExpress | Fuzzy significance scoring para varianzas presupuestales — priorizar lo realmente importante | `inference/fuzzy` (Sugeno) | Low | Si |
| 4.3 | **Win-Rate Estimation** | Mejora PropuestaYa | Bayesian P(win \| client, pricing, competition) — calibrar esfuerzo por oportunidad | `inference/bayesian` | Medium | Si |
| 4.4 | **ComplianceGuard** | Nuevo vertical | Compliance regulatorio colombiano (DIAN, SIC) con reglas clasicas + Bayes riesgo de auditoria | `inference/classical` + `inference/bayesian` | Medium | Si |
| 4.5 | **HRScreen** | Nuevo vertical | CV screening con fuzzy scoring + reglas clasicas de fairness — ranking explicable | `inference/fuzzy` + `inference/classical` | Low-Medium | Si |
| 4.6 | **ProcureScore** | Nuevo vertical | Comparacion de cotizaciones: LLM extrae datos + fuzzy multi-criterio evalua | `inference/fuzzy` | Low | Si |
| 4.7 | **CostGuard** | Mejora plataforma | Controlador fuzzy de presupuesto de tokens — degradacion gradual en vez de corte duro | `inference/fuzzy` (Mamdani) | Low | Si |
| 4.8 | **ReliabilityNet** | Mejora plataforma | Bayesian P(success \| conditions) para predecir confiabilidad de Skills y auto-escalar | `inference/bayesian` | Medium | Si |

**Detalles:**

- **Reconciliation Engine** — CierreExpress reconcilia transacciones bancarias con registros
  contables. Hoy el LLM hace todo el matching. Con reglas clasicas: matching exacto por
  monto+fecha+referencia (deterministico, gratis, instantaneo). Solo los que no matchean
  van al LLM. Reduccion estimada de 60-80% en API calls. Las reglas de matching son
  auditables y configurables por cliente.

- **Variance Analysis** — Las varianzas presupuestales son ruidosas. Fuzzy Sugeno clasifica
  cada varianza: magnitud (trivial, notable, critica) x frecuencia (aislada, recurrente) x
  tendencia (mejorando, estable, empeorando). Output: score numerico ponderado que permite
  al contador priorizar. "Esta varianza de $50K es critica-recurrente-empeorando (score 0.92)
  vs esta de $80K que es notable-aislada-mejorando (score 0.34)."

- **Win-Rate Estimation** — PropuestaYa genera propuestas comerciales. Con una red bayesiana:
  P(ganar | tipo_cliente, rango_precio, competencia_detectada, historial_similar). Calibra
  cuanto esfuerzo invertir en cada oportunidad. Se alimenta con datos historicos del usuario.
  Explicacion bayesiana: "la probabilidad sube de 40% a 65% porque el cliente es recurrente".

- **ComplianceGuard** — Vertical de compliance para PyMEs colombianas. Reglas clasicas
  codifican requisitos regulatorios (DIAN: facturacion electronica, SIC: proteccion de datos,
  UGPP: seguridad social). Bayes calcula P(auditoria | sector, tamano, historico). Forward
  chaining: "dados estos hechos de tu empresa, estos son los requisitos que no cumples".
  Traza completa para evidencia.

- **HRScreen** — Screening de CVs con fairness explicable. Fuzzy scoring: experiencia
  (junior, mid, senior), educacion (relevancia baja/media/alta), skills (parcial/completo).
  Reglas clasicas de fairness: no penalizar gaps, no filtrar por universidad, edad no es
  variable. Output: ranking numerico + explicacion de cada score. Auditable por diseno.

- **ProcureScore** — Comparacion de cotizaciones para compras. LLM extrae datos
  estructurados de PDFs/emails (proveedor, precio, plazo, condiciones). Fuzzy multi-criterio
  evalua: precio (competitivo, promedio, caro), plazo (urgente, normal, holgado),
  condiciones (favorables, estandar, restrictivas). Ranking automatico con explicacion.

- **CostGuard** — Aluna consume tokens de Claude API. En vez de un corte duro ("presupuesto
  agotado, stop"), fuzzy Mamdani implementa degradacion gradual. Variables: porcentaje_usado
  (bajo, medio, alto, critico), prioridad_tarea (baja, normal, alta, urgente). Reglas:
  "si porcentaje es alto Y prioridad es baja ENTONCES restriccion es fuerte". El agente
  ajusta su comportamiento (respuestas mas cortas, menos tool calls) antes de agotar el
  presupuesto.

- **ReliabilityNet** — Red bayesiana que modela P(exito_skill | hora_del_dia, carga_actual,
  tipo_input, historial_reciente). Si la probabilidad de exito cae (ej: un API externo
  esta inestable), la plataforma puede: reintentar con backoff, escalar a otro Skill,
  o notificar al usuario antes de fallar. Observability predictiva, no solo reactiva.
