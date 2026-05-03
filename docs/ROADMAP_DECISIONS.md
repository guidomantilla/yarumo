# Roadmap — sdks/decisions/

Futuro del ecosistema de decisiones: companions, ciclo de vida, DaaS.

---

## Estado actual

`sdks/decisions/core/` — 99.1% coverage, 0 linter issues. 6 paquetes:

- `evaluate/` — servicio unificado con 6 paradigmas + cascade pipeline
- `schema/` — tipos serializables (RuleSet, DeductiveConfig, BayesianConfig, FuzzyConfig, TableConfig, ScorecardConfig, TreeConfig)
- `adapters/` — 4 binder interfaces segregadas + composite `Binder[D]`
- `explain/` — 6 explainer interfaces segregadas
- `validate/` — 6 métodos de validación por paradigma
- `repository/` — interface de persistencia (sin implementación)

`core/` es el motor puro. Evalúa, valida, explica. No tiene ciclo de vida, no tiene persistencia, no tiene HTTP.

---

## `lifecycle/` — ciclo de vida de reglas

Lo que falta en `sdks/decisions/` para manejar reglas como artefactos versionados:

| Responsabilidad | Qué hace |
|----------------|----------|
| Estados | draft → validated → published → active → deprecated → archived |
| Versionamiento | Semver de rulesets, inmutabilidad de versiones publicadas |
| Vinculación | Ruleset X usa ontología `veterinaria:v1.0` |
| Migración | Cuando la ontología cambia, detectar reglas afectadas, proponer cambios |
| Auditoría | Quién cambió qué, cuándo, por qué |

`core/` no necesita saber nada de esto. `lifecycle/` **usa** `core/validate/` para validar antes de publicar, pero `core/` no sabe que `lifecycle/` existe.

```
lifecycle/ → core/validate/   (para validar antes de publicar)
lifecycle/ → core/schema/     (para leer rulesets)
lifecycle/ → ontology SDK     (para verificar compatibilidad con versión de ontología)
```

---

## SDK Companion Modules

Regla fundamental: **solo crear módulos cuando hay un consumidor real.** El DaaS es el primer consumidor. Todo se implementa dentro de `apps/daas/internal/` primero. Se extrae al SDK como companion module cuando haya un segundo consumidor.

```
sdks/decisions/
    core/                       ← orquestación genérica (existe hoy)

    lifecycle/                  ← ciclo de vida de reglas (por diseñar)

    storage/                    ← agrupador
        postgres/               ← Repository + evaluate.Log con pgx
        redis/                  ← Repository (cache) con go-redis
        mongo/                  ← Repository + evaluate.Log con mongo-driver
        filesystem/             ← Repository con YAML/JSON files

    socratic/                   ← desambiguación NL → reglas (licencia propia)

    llm/                        ← LLMClient interface + LLMExplainer

    embeddings/                 ← Semantic search sobre rulesets

    telemetry/                  ← agrupador
        otel/                   ← ObservableService con OTel spans + metrics

    endpoints/                  ← agrupador
        http/                   ← REST handlers (evaluate, validate, CRUD)
        grpc/                   ← Proto + server/client
```

### `storage/{datasource}/`

Cada datasource implementa `Repository` y/o `evaluate.Log` del core.

| Módulo | Abstracciones | Deps |
|--------|--------------|------|
| `storage/postgres/` | `PostgresRepository`, `PostgresLog` | core + pgx/v5 |
| `storage/redis/` | `RedisRepository` (cache read-through decorator) | core + go-redis/v9 |
| `storage/mongo/` | `MongoRepository`, `MongoLog` | core + mongo-driver/v2 |
| `storage/filesystem/` | `FilesystemRepository` (YAML/JSON files) | core (solo stdlib) |

### `llm/`

| Abstracción | Responsabilidad |
|-------------|-----------------|
| `LLMClient` | Interface genérica de LLM (`Complete(ctx, prompt) (string, error)`) |
| `LLMExplainer` | `Explainer` que enhances templates via `LLMClient`, graceful degradation al template base si falla |

### `socratic/`

Módulo socrático de desambiguación. Recibe lenguaje natural, usa LLM como parser (texto → JSON), valida contra ontología de dominio con código Go determinístico, genera preguntas de desambiguación, y produce salida serializable (JSON/YAML/TOML) con reglas formales sin ambigüedades.

**No depende de `core/` en código.** Su salida es `[]byte` serializado. El DaaS hace la traducción a `schema.RuleSet`.

**Licencia propia** — no Apache 2.0. Es el diferenciador del producto.

Ver `docs/ONTOLOGIES_SOCRATIC.md` para arquitectura detallada.

### `embeddings/`

| Abstracción | Responsabilidad |
|-------------|-----------------|
| `EmbeddingProvider` | Genera embeddings de texto (interface genérica) |
| `VectorStore` | Almacena y busca vectores por similaridad (pgvector, pinecone, en memoria) |
| `SemanticIndex` | Indexa rulesets al persistirlos, busca por significado |

### `telemetry/otel/`

| Abstracción | Responsabilidad |
|-------------|-----------------|
| `ObservableService[D]` | Decorator sobre `Service[D]` con spans + metrics (decision_count, decision_duration, decision_errors) |

### `endpoints/{http,grpc}/`

| Módulo | Abstracciones |
|--------|--------------|
| `endpoints/http/` | `EvaluateHandler`, `ValidateHandler`, `RulesetHandler`, `Mount(mux)` — handlers puros `http.Handler` |
| `endpoints/grpc/` | `DecisionServer`, `DecisionClient`, `decision.proto` |

### Prioridades de extracción

| Prioridad | Módulo | Trigger |
|-----------|--------|---------|
| Alta | `storage/postgres/` | Cuando haya segundo consumidor de persistencia |
| Alta | `socratic/` | Parte del DaaS desde el inicio (licencia propia) |
| Alta | `llm/` | Cuando haya segundo consumidor de LLM explanations |
| Media | `telemetry/otel/` | Cuando haya segundo consumidor de observabilidad |
| Media | `endpoints/http/` | Cuando alguien quiera exponer el engine como REST genérico |
| Baja | `storage/redis/` | Cache de rulesets |
| Baja | `storage/filesystem/` | Rulesets en disco (CLI, dev) |
| Baja | `storage/mongo/` | Consumidor con MongoDB |
| Baja | `endpoints/grpc/` | Consumidor gRPC |
| Baja | `embeddings/` | Repositorios grandes con semantic search |

---

## DaaS — Producto (apps/daas/)

Primer consumidor real del SDK. El usuario final es una **persona no técnica** (ej. PYME LATAM) que escribe reglas de negocio en lenguaje natural. El sistema desambigua, formaliza, y evalúa — el usuario nunca ve JSON ni escribe código.

```
┌─────────────────────────────────────────────────────────────┐
│                        DaaS (producto SaaS)                  │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │   Frontend    │───▶│  Backend API  │───▶│  PostgreSQL   │   │
│  │   (SPA)      │    │  (Go)        │    │              │   │
│  └──────────────┘    └──────┬───────┘    └──────────────┘   │
│                             │                                │
│                    ┌────────┴────────┐                       │
│                    │                 │                        │
│              ┌─────▼─────┐   ┌──────▼──────┐               │
│              │ Socrático  │   │  Decision   │               │
│              │ (NL→regla) │   │  Engine SDK │               │
│              └─────┬──────┘   └─────────────┘               │
│                    │                                         │
│              ┌─────▼─────┐                                  │
│              │    LLM     │                                  │
│              │ (externo)  │                                  │
│              └────────────┘                                  │
└─────────────────────────────────────────────────────────────┘
```

Flujo del usuario no técnico:
1. Escribe en lenguaje natural: "los clientes morosos no pueden comprar a crédito"
2. El socrático (LLM + código Go) desambigua: "¿qué es moroso? ¿más de 30 días? ¿más de 60?"
3. El usuario responde hasta que no hay ambigüedades
4. El sistema traduce a regla formal y la persiste
5. La regla se evalúa contra datos reales via API o UI

### Frontend (SPA)

Seis áreas funcionales:

- **NL Authoring (socrático)** — campo de texto libre, desambiguación interactiva, el usuario nunca ve JSON
- **Tech Authoring** — editor por paradigma para usuarios técnicos (deductive, bayesian, fuzzy, table, scorecard, tree)
- **Decision Playground** — seleccionar ruleset, ingresar inputs, ejecutar, ver resultado + explanation
- **Cascade Builder** — armar pipelines multi-paradigma visualmente, ver resultado stage-by-stage
- **Audit Explorer** — tabla paginada de decisiones pasadas, filtros, detalle, export
- **Validation Dashboard** — subir ruleset, ver reporte completo (contradicciones, redundancias, gaps)

### Generic Binder

El DaaS es domain-agnostic. Un solo binder implementa las 4 interfaces:

```go
type DecisionInput struct {
    Facts    map[string]bool    `json:"facts,omitempty"`    // Deductive
    Evidence map[string]string  `json:"evidence,omitempty"` // Bayesian
    Inputs   map[string]float64 `json:"inputs,omitempty"`   // Fuzzy
    Data     map[string]any     `json:"data,omitempty"`     // Table, Scorecard, Tree
}

type PassthroughBinder struct{}
var _ evaluate.Binder[DecisionInput] = PassthroughBinder{}
```

### REST API

Base: `/api/v1`

| Endpoint | Método | Descripción |
|----------|--------|-------------|
| `/api/v1/evaluate` | POST | Evaluar (6 paradigmas, campo `paradigm` determina engine) |
| `/api/v1/evaluate/cascade` | POST | Evaluar pipeline multi-paradigma |
| `/api/v1/validate` | POST | Validar ruleset (detecta configs presentes, ejecuta validadores) |
| `/api/v1/rulesets` | GET | Listar rulesets |
| `/api/v1/rulesets/{name}/{version}` | GET | Obtener ruleset |
| `/api/v1/rulesets` | POST | Crear (valida antes de guardar, 422 si inválido) |
| `/api/v1/rulesets/{name}/{version}` | PUT | Actualizar (valida antes de guardar) |
| `/api/v1/rulesets/{name}/{version}` | DELETE | Eliminar (RESTRICT si tiene audit entries) |
| `/api/v1/audit` | GET | Listar audit trail (paginado, filtrable) |
| `/api/v1/audit/{id}` | GET | Detalle de audit entry |
| `/health` | GET | Liveness (siempre 200) |
| `/ready` | GET | Readiness (verifica DB) |

Converters nombrados (built-in) para cascade:
- `deductive_to_bayesian`, `deductive_to_fuzzy`
- `bayesian_to_deductive`, `bayesian_to_fuzzy`
- `fuzzy_to_deductive`, `fuzzy_to_bayesian`

### Storage (PostgreSQL)

```sql
CREATE TABLE IF NOT EXISTS decision_rulesets (
    name        TEXT        NOT NULL,
    version     TEXT        NOT NULL,
    config      JSONB       NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (name, version)
);

CREATE TABLE IF NOT EXISTS decision_audit (
    id              TEXT        PRIMARY KEY,
    timestamp       TIMESTAMPTZ NOT NULL,
    ruleset_name    TEXT        NOT NULL,
    ruleset_version TEXT        NOT NULL,
    paradigm        TEXT        NOT NULL,
    request         JSONB       NOT NULL,
    result          JSONB       NOT NULL,
    explanation     TEXT        NOT NULL,
    duration_ms     BIGINT      NOT NULL,
    metadata        JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Sin FK entre ellas (audit entries son inmutables, independientes del ciclo de vida del ruleset).

### Estructura del módulo

```
apps/daas/
    go.mod
    .golangci.yml
    .testcoverage.yml
    CODING_STANDARDS.md

    main.go                   -- lifecycle: config → telemetry → resources → wiring → managed → signal wait
    resources.go              -- pgxpool creation
    wiring.go                 -- assemble services, repo, handlers, mux
    types.go                  -- DecisionInput, converters registry
    binder.go                 -- PassthroughBinder

    internal/
        storage/
            types.go          -- QueryableAuditLog interface, AuditFilter
            errors.go         -- StorageType, Error, sentinels
            repository.go     -- postgresRepository
            audit.go          -- postgresAuditLog
            options.go        -- WithSchema, WithTablePrefix
            migrations.go     -- embedded SQL strings

        api/
            types.go          -- DTOs: EvaluateRequest/Response, CascadeRequest, ValidationResponse
            errors.go         -- HandlerType, Error, sentinels
            evaluate.go       -- POST /api/v1/evaluate
            cascade.go        -- POST /api/v1/evaluate/cascade
            validate.go       -- POST /api/v1/validate
            rulesets.go       -- CRUD /api/v1/rulesets
            audit.go          -- GET /api/v1/audit
            health.go         -- /health, /ready
            routes.go         -- Mount(mux, deps)
            middleware.go     -- requestID, logging, recovery, content-type

        converters/
            converters.go     -- named StageConverter registry
```

### Deployment

```
Docker Compose: daas (:8080) + postgres (:5432)
- Backend sirve la SPA como static files (o reverse proxy en dev)
- Postgres con migrations al startup
- Producción: separar frontend (CDN) y backend (container)
```

### Variables de entorno

| Variable | Default | Descripción |
|---|---|---|
| `DATABASE_URL` | (requerido) | PostgreSQL connection string |
| `HOST` | `0.0.0.0` | Bind address |
| `PORT` | `8080` | Puerto HTTP |
| `LOG_LEVEL` | `info` | Nivel de log |
| `EXPLAINER_LOCALE` | `es` | Locale para explicaciones (es/en) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | — | Collector endpoint (opcional) |

### Decisiones de diseño

| Decisión | Razón |
|---|---|
| Todo en `apps/daas/`, no companion modules | Extraer al SDK cuando haya segundo consumidor |
| Un solo service unificado | `evaluate.Service[DecisionInput]` maneja los 6 paradigmas |
| PassthroughBinder | Binding es siempre app-specific |
| JSONB para rulesets | Flexible, query-able, evita tablas por paradigma |
| Sin FK audit→rulesets | Audit entries son inmutables; simplifica deletes |
| `http.ServeMux` stdlib | Consistente con `common/http.Server` |
| Named converters (no custom code) | Seguridad — no ejecutar código arbitrario del cliente |

4 fases de implementación.

### Lo que el DaaS NO hace (MVP)

| Responsabilidad | Dueño | Razón |
|---|---|---|
| Binders de dominio | Apps consumidoras | Binding es inherentemente domain-specific |
| Autenticación/autorización | Infraestructura (reverse proxy) | Cross-cutting, no decision-specific |
| LLM explanations | Futuro companion module | Not MVP |
| NL rule authoring (socrático) | `sdks/decisions/socratic/` companion | Parte del producto, no del MVP API |
| gRPC endpoints | Futuro | No hay consumidor |
| Multi-tenancy | Futuro | Single-tenant para MVP |

### Fases de implementación

**Phase 0 — Scaffold**
- `apps/daas/` con go.mod, linter config, coding standards
- Registrar en `go.work`
- Estructura de directorios vacía

**Phase 1 — Storage**
- `internal/storage/` — PostgresRepository + PostgresAuditLog
- Migrations embebidas
- Tests con pgxmock o testcontainers

**Phase 2 — API Handlers**
- `internal/api/` — todos los endpoints
- DTOs, middleware, routes
- Un solo endpoint `/api/v1/evaluate` para los 6 paradigmas
- Validación para los 6 paradigmas
- Tests con httptest

**Phase 3 — App Wiring**
- `main.go`, `resources.go`, `wiring.go`
- `binder.go` (PassthroughBinder), `types.go`
- Converters registry
- Docker Compose (postgres + app)

**Phase 4 — Hardening**
- Request validation exhaustiva
- Paginación en list endpoints
- Migration runner al startup
- Correlation IDs en logs
- Integration tests end-to-end

### DaaS API Layer (post-MVP)

Los `endpoints/{http,grpc}/` del SDK son handlers genéricos. La capa de producto necesita:

- Multi-tenancy (tenant isolation, data partitioning)
- Auth (API keys, OAuth2, JWT)
- Rate limiting
- Usage metering / billing
- API versioning
- Deployment (containers, CI/CD)

---

## Vertical Orchestrators

`CascadePipeline` existe en core/, pero cada vertical tiene su propia cascada:

| Producto | Cascada |
|---|---|
| LendingBrain | PreCheck → KnockOut → Score → Risk → Product → Sensitivity |
| UnderwriteIQ | PreCheck → KnockOut → Risk → Rate → Reinsurance → Sensitivity |
| GovRules | Docs → Elegibilidad → Vulnerabilidad → Priorización → Plazos → Motivación |
| ClinicalRules | Triage → Dx → Exámenes → Tratamiento → Disposición → Seguimiento |
| AgroDecide | Monitoreo → Riesgo → Estado → Acción → Priorización → Registro |
| ComplianceEngine | Identificar → Evaluar → Controlar → Monitorear → Reportar |

Falta un framework o patrón para definir estas cascadas verticales sobre CascadePipeline.

---

## Vista por capas (estado general)

```
┌─────────────────────────────────────────────────────────┐
│  FRONTENDS (0% construido)                              │
│  Dashboard compliance, portal analista, portal ciudadano│
│  app movil agro, review UI KnowledgeForge              │
├─────────────────────────────────────────────────────────┤
│  CHANNELS (0% construido)                               │
│  WhatsApp Business API, SMS gateway, email/Slack        │
├─────────────────────────────────────────────────────────┤
│  AGENT RUNTIME (0% construido)                          │
│  THINK→ACT→OBSERVE loop, escalation, guardrails        │
├─────────────────────────────────────────────────────────┤
│  DaaS API LAYER (0% construido)                         │
│  Multi-tenancy, auth, rate limiting, API keys           │
├─────────────────────────────────────────────────────────┤
│  VERTICAL ORCHESTRATORS (0% construido)                 │
│  Cascadas específicas por producto                      │
├─────────────────────────────────────────────────────────┤
│  CROSS-CUTTING SERVICES (0% construido)                 │
│  Notifications, calendar, OCR, FHIR, batch             │
├─────────────────────────────────────────────────────────┤
│  KNOWLEDGE FORGE (0% construido)                        │
│  Ingestion, extraction, review, coverage, versioning   │
├─────────────────────────────────────────────────────────┤
│  SDK COMPANIONS (0% construido, arquitectura definida)  │
│  storage/, llm/, nlp/, embeddings/, telemetry/, endpts/ │
├─────────────────────────────────────────────────────────┤
│  ✅ SDK CORE (100% construido)                          │
│  evaluate, validate, explain, repository               │
├─────────────────────────────────────────────────────────┤
│  ✅ INFERENCE (100% construido, 5 paradigmas)           │
│  deductive, bayesian, fuzzy, causal, mcdm             │
├─────────────────────────────────────────────────────────┤
│  ✅ MATHS (100% construido, gaps en stats/)            │
│  logic, fuzzy, sets, stats (ver ROADMAP_MATH.md)       │
└─────────────────────────────────────────────────────────┘
```

---

## Prioridad de construcción

Orden bottom-up — cada capa habilita las de arriba:

| Paso | Qué | Estado |
|------|-----|--------|
| 1 | Inference extensions | ✅ COMPLETADO |
| 2 | SDK companions (storage/postgres, llm/, endpoints/http) | Arquitectura definida, código por escribir |
| 3 | DaaS API layer (multi-tenancy, auth, rate limiting) | Necesita endpoints/ primero |
| 4 | KnowledgeForge (ingestion, extraction, review) | Necesita llm/ + nlp/ primero |
| 5 | Cross-cutting services (notifications, calendar, batch) | Según la vertical que se ataque primero |
| 6 | Vertical orchestrators + Frontends | Específicos al primer producto |
| 7 | Channels (WhatsApp, SMS) + Agent runtime | Cuando se active la capa agéntica |

**La decisión crítica**: cuál producto llevar a mercado primero determina qué capas construir primero.
