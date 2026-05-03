El ciclo de vida natural sería:

## 1. Seed (Semilla)

Alguien crea una ontología inicial. Tres caminos posibles:
- **Manual** — un experto de dominio escribe YAML/JSON
- **Desde documentos** — KnowledgeForge extrae entidades/campos de PDFs/ISOs/regulaciones y propone un borrador
- **Desde conversación** — Socrático va descubriendo entidades mientras el usuario describe reglas (el problema de circularidad)

---

## RDF/OWL vs. Lo que Yarumo necesita

| Estándar | Para qué sirve | Por qué no encaja |
|----------|----------------|-------------------|
| RDF/RDFS | Grafos de conocimiento abiertos, linked data en la web | Yarumo no publica datos en la web semántica |
| OWL | Razonamiento lógico sobre clases (subsumption, clasificación) | Ya tenemos engines de razonamiento propios (deductive, bayesian, etc.) |
| SPARQL | Consultar grafos RDF | No tenemos triple stores ni necesitamos uno |
| SHACL | Validar formas de grafos RDF | Útil en concepto, pero atado al ecosistema RDF |
| JSON-LD | RDF serializado como JSON | Complejidad innecesaria si no usamos RDF |
| SKOS | Taxonomías y vocabularios controlados | El concepto es útil, pero no necesitamos el formato |

### El problema de RDF/OWL

Son estándares diseñados para el **Open World Assumption** — "lo que no sé, podría ser verdad". Yarumo opera en **Closed World** — "lo que no está en la ontología, no existe". Son filosofías opuestas.

Además, meter RDF/OWL implica:
- Dependencias pesadas (triple stores, reasoners externos)
- Complejidad de serialización enorme
- Curva de aprendizaje brutal para los usuarios del producto

### Lo que sí necesitamos

Las ontologías de Yarumo son **dominios acotados y cerrados**. Un veterinario define perro, peso, raza. Un banco define cliente, ingreso, score. No necesitan enlazarse con DBpedia.

Lo que necesitamos es más parecido a un **schema tipado**:

```yaml
# Ejemplo: ontología de veterinaria v1
domain: veterinaria
entities:
  perro:
    fields:
      peso:
        type: number
        unit: kg
        range: [0.5, 120]
      raza:
        type: enum
        values: [labrador, pastor_aleman, bulldog, ...]
      edad:
        type: number
        unit: años
        range: [0, 25]
    relationships:
      dueño: { entity: cliente, cardinality: many-to-one }
  cliente:
    fields:
      nombre:
        type: string
      telefono:
        type: string
        pattern: "^\\+?[0-9]{7,15}$"
```

### Sobre la captura de texto

En la etapa seed, la entrada puede ser:

1. **YAML/JSON directo** — un técnico lo escribe a mano
2. **Texto libre → LLM → YAML** — KnowledgeForge o Socrático parsean un documento y proponen la estructura
3. **UI guiada** — el DaaS ofrece un formulario (futuro)

Pero el **artefacto resultante** siempre es estructurado (YAML/JSON), nunca texto libre. El texto libre es *input*, no *estado*.

---

¿Tomamos los conceptos útiles de esos estándares? Sí:
- De **SHACL**: la idea de constraints sobre campos
- De **SKOS**: sinónimos y labels multilingüe
- De **OWL**: herencia entre entidades (un `pastor_aleman` es un `perro`)

Pero como **tipos Go propios**, no como dependencias de la web semántica.

---

## 2. Validate (Validación de la ontología misma)
La ontología como artefacto necesita validarse:
- Campos tienen tipos válidos, rangos coherentes
- No hay entidades huérfanas ni relaciones rotas
- Constraints no se contradicen

La validación tiene dos capas: el **compilador** (estructura) y el **reviewer** (sentido de negocio).

### Compilador — Go determinístico, bloquea si falla

Un paquete `validate` dentro del SDK de ontologías. Recibe el artefacto parseado, retorna `[]ValidationError`. Si hay errores, no se puede publicar.

```
YAML/JSON → Loader → Ontology struct → Validate() → []ValidationError
                                                      └─ vacío = válida, puede publicar
```

Qué valida:

**Integridad de tipos:**
- `type: number` → `range` debe ser `[min, max]` numérico
- `type: enum` → `values` no puede estar vacío
- `type: string` + `pattern` → el regex debe compilar

**Integridad de relaciones:**
- Si `perro.relationships.dueño` apunta a `entity: cliente`, `cliente` debe existir
- Cardinalidades válidas (`one-to-one`, `one-to-many`, `many-to-one`, `many-to-many`)

**Coherencia de constraints:**
- `range: [120, 0.5]` → min > max, inválido
- Dos campos con el mismo nombre en la misma entidad → duplicado
- Un `enum` con valores duplicados

**Completitud:**
- Entidad sin campos → warning o error
- Campo con `unit` pero sin `range` → warning (¿tiene sentido una unidad sin rango?)

No hay LLM en este paso. Es puro Go, puro determinismo.

### Reviewer — LLM + datos, sugiere pero no bloquea

El compilador verifica estructura, pero **no puede validar sentido de negocio**. Puede decir que `range: [0.5, 120]` es válido (min < max, ambos numéricos), pero no puede decir si 120 kg para un perro tiene sentido.

Dos formas de asistir al humano:

1. **LLM como reviewer** — "Estás diciendo que un perro puede pesar 120 kg. Las razas más grandes llegan a ~90 kg. ¿Estás seguro?" — Es una sugerencia, no una decisión. El humano acepta o rechaza.

2. **Datos históricos** — Si ya existen datos del dominio, se puede comparar: "El 99.9% de los perros registrados pesan menos de 80 kg, pero tu rango llega a 120 kg." — Estadística, no opinión.

### El flujo real

```
Seed (texto) → Compilador (estructura) → Reviewer (sentido) → Humano aprueba
                    │                         │
                    │ Go determinístico        │ LLM + datos (asistencia)
                    │ BLOQUEA si falla         │ SUGIERE, no bloquea
```

El compilador es **gate** (no pasa si falla). El reviewer es **advisory** (sugiere, el humano decide).

---

## 3. Publish (Publicación)

La ontología pasa de borrador a "activa". Desde ese momento:
- Socrático la usa para disambiguar NL → reglas
- El runtime validator la usa para rechazar output de LLMs
- KnowledgeForge la usa para clasificar reglas extraídas

Publicar implica versionar. Siempre.

### Por qué

Una ontología publicada tiene **consumidores activos**: reglas del Socrático, configuraciones del Decision Engine, validaciones de agentes. Si cambias la ontología sin versionar:

- Una regla dice `peso >= 25` pero el campo `peso` ya no existe (lo renombraron a `peso_kg`) → **regla rota**
- Un agente valida contra una ontología que ya cambió → **comportamiento inconsistente**
- No puedes reproducir por qué una decisión se tomó hace 3 meses → **auditoría imposible**

### Implicaciones

**Cada publicación crea una versión inmutable.** Una vez publicada, esa versión no se modifica — se crea una nueva.

```
v1.0 (publicada 2026-01-15) → 3 rulesets la usan
v1.1 (publicada 2026-02-03) → 1 ruleset nuevo la usa
v2.0 (publicada 2026-03-01) → breaking change, nadie la usa aún
```

**Las reglas están vinculadas a una versión específica:**
- RuleSet X fue creado con ontología `veterinaria:v1.0`
- Si la ontología evoluciona a `v2.0`, RuleSet X sigue usando `v1.0` hasta que alguien lo migre

**Breaking vs non-breaking:**

| Cambio | Tipo | Ejemplo |
|--------|------|---------|
| Agregar entidad | Non-breaking | Nueva entidad `gato` |
| Agregar campo opcional | Non-breaking | `perro.chip_id` (opcional) |
| Ampliar rango | Non-breaking | `peso: [0.5, 120]` → `[0.5, 150]` |
| Renombrar campo | **Breaking** | `peso` → `peso_kg` |
| Eliminar entidad | **Breaking** | Borrar `cliente` |
| Reducir rango | **Breaking** | `peso: [0.5, 120]` → `[1, 80]` — reglas con `peso >= 0.7` quedan fuera |
| Cambiar tipo | **Breaking** | `edad: number` → `edad: enum` |

**Semver simplificado:**
- `MAJOR.MINOR` — major para breaking changes, minor para non-breaking
- No necesitamos patch — una ontología no tiene "bugs", tiene cambios de dominio

**Quién decide publicar:**
- El humano. Siempre. Después de que el compilador y el reviewer pasaron.
- Publicar es un acto explícito, no automático.

### Impacto en reglas existentes

Las reglas nacen vinculadas a una versión de ontología. El impacto de una nueva versión depende del tipo de cambio:

#### Non-breaking (v1.0 → v1.1)

Las 5 reglas existentes siguen funcionando sin cambio. La nueva versión solo agrega cosas (nueva entidad, campo opcional, rango más amplio). Las reglas no referencian lo nuevo, así que no les afecta.

```
Ontología v1.0 → Regla 1, 2, 3, 4, 5 ✓ (siguen válidas)
Ontología v1.1 → Regla 1, 2, 3, 4, 5 ✓ (siguen válidas, pueden migrar opcionalmente)
               → Regla 6 (nueva, usa el campo nuevo)
```

Opción: las reglas pueden migrar a v1.1 automáticamente (es compatible). O pueden quedarse en v1.0 hasta que alguien decida.

#### Breaking (v1.x → v2.0)

Aquí es donde duele. Si `peso` se renombró a `peso_kg`:

```
Ontología v2.0 → Regla 1 usa "peso >= 25" → ❌ campo "peso" no existe en v2.0
               → Regla 2 usa "raza == labrador" → ✓ no afectada
               → Regla 3 usa "peso < 10" → ❌ mismo problema
```

Tres estrategias posibles:

| Estrategia | Cómo funciona | Trade-off |
|-----------|---------------|-----------|
| Congelar | Las reglas se quedan en v1.0, las nuevas usan v2.0. Coexisten versiones | Simple, pero fragmenta el sistema |
| Migrar | Detectar reglas afectadas, proponer cambios (`peso` → `peso_kg`), humano aprueba | Limpio, pero requiere tooling de migración |
| Romper | Publicar v2.0, las reglas incompatibles se marcan como "requiere revisión" y se desactivan | Fuerza limpieza, pero puede romper producción |

#### Lo que probablemente necesitamos

**Migrar** es la respuesta correcta. El flujo sería:

```
Ontología v2.0 (borrador)
    │
    ▼
Diff contra v1.x → cambios detectados: peso → peso_kg
    │
    ▼
Buscar reglas que usan "peso" → Regla 1, Regla 3
    │
    ▼
Proponer migración: "peso" → "peso_kg" en Regla 1 y Regla 3
    │
    ▼
Humano aprueba → reglas migradas → ontología v2.0 publicada
```

Esto es el paso **6. Migrate** del ciclo de vida. Publicar una versión breaking **no debería ser posible** sin resolver las reglas afectadas primero.

**Versiones viejas:** cuando ninguna regla referencia una versión (todas migraron), se marca como deprecated y eventualmente se archiva. Es mantenimiento operacional, no un paso del ciclo de vida.

---

## 4. Consume (Consumo)

La ontología publicada se consume en tres momentos distintos respecto al LLM:

### Antes del LLM — constraints en prompts

Antes de enviar un prompt al LLM, se genera un JSON Schema o instrucciones desde la ontología: "las entidades válidas son X, Y, Z, los campos de perro son peso (number, 0.5-120 kg), raza (enum: labrador, bulldog...)". Esto reduce alucinaciones **antes** de que ocurran.

### Durante — Socrático (desambiguación)

Cuando un usuario dice "los perros grandes necesitan bozal", el Socrático consulta la ontología: ¿existe la entidad `perro`? ¿tiene un campo que pueda significar "grande"? ¿`peso` tiene rango? Si el usuario dice "150 kg", el Socrático sabe que está fuera de rango `[0.5, 120]` y pregunta.

### Después del LLM — validación de output

Cualquier agente/LLM que produzca output con datos del dominio, se valida contra la ontología: "el agente dijo que el perro pesa 200 kg → la ontología dice máximo 120 → rechazar/marcar".

```
Constraints (antes) → LLM → Validación (después)
       │                          │
       └── ontología ─────────────┘
                │
                └── Socrático (durante)
```

### Enriquecimiento — efecto secundario del consumo

El enriquecimiento no es un paso separado del ciclo de vida. Es lo que pasa cuando el consumo detecta algo nuevo:

- El Socrático consume la ontología y el usuario dice "el perro tiene chip GPS" → `chip_gps` no existe → el Socrático propone: "este campo no existe, ¿quieres agregarlo?"
- KnowledgeForge procesa un documento nuevo y encuentra un campo que no está en la ontología → propone extensión
- Un admin detecta manualmente que falta algo → lo agrega

En los tres casos, la propuesta de agregar **re-entra al ciclo como un nuevo Seed** → Validate → Publish (nueva versión minor).

```
Consume → detecta algo nuevo → Seed (propuesta) → Validate → Publish (v1.1)
```

No es un paso aparte — es el ciclo alimentándose a sí mismo.

---

## Arquitectura propuesta — OntologyRegistry

Tres capas, mismo patrón que Decision Engine (apps/daas → sdks/decisions/core → modules/common):

```
apps/ontology-registry/    ← producto desplegable
├── api/                      endpoints REST/gRPC
├── storage/                  persistencia (DB)
├── workers/                  jobs async (review LLM, migration proposals)
└── ui/                       frontend (futuro)

sdks/ontology/             ← lógica de negocio (library), licencia propia
├── registry/                 publicar, versionar, consultar
├── migrate/                  proponer migraciones de reglas
├── review/                   LLM reviewer (sentido de negocio)
├── constraints/              generar constraints para prompts
└── runtime/                  validar output de agentes en runtime

modules/ontology/          ← primitivas puras (library), Apache 2.0
├── schema/                   tipos: Entity, Field, Constraint, Range, Unit, Relationship
├── loader/                   YAML/JSON → Ontology struct
├── validate/                 compilador (estructura)
└── diff/                     comparar versiones
```

### Criterio de separación

**modules/** — ¿Puedo usarlo sin saber qué es una ontología de negocio? Son tipos de datos puros, carga de archivos, validación estructural, diff de versiones. Genérico y reutilizable. Apache 2.0.

**sdks/** — ¿Necesito entender el ciclo de vida, reglas, o LLMs? Publicación, versionamiento, migración de reglas, review con LLM, validación de agentes en runtime. Lógica de negocio. Licencia propia.

**apps/** — ¿Necesito exponer como servicio? API HTTP, persistencia en DB, jobs async, autenticación, tenancy, permisos, UI. Producto desplegable.

Cada capa solo conoce la de abajo. El app no sabe de lógica de diff. El SDK no sabe de HTTP ni de DB. El module no sabe de ontologías de negocio.