## Origen de la idea

Conversación exploratoria sobre ontologías y agentes de IA:

1. Qué es una ontología y cómo luce (YAML, diagrama)
2. En la realidad qué tan cierto es que sirve (el LLM siempre puede alucinar)
3. Cómo valida código determinístico contra algo que es "mero texto .md"
4. Quién transforma el .md a YAML y si el código generado requiere deploy
   por cada cambio
5. Cómo hacer que el LLM sea siempre socrático — **respuesta: no confías en
   el LLM, el código Go es el socrático**
6. Si se puede tener un módulo genérico en Go para el socratismo
7. Cómo crear el prompt para el agente LLM extractor
8. Cómo pedirle al agente de Yarumo que diseñe todo esto

Insight clave: el LLM no es confiable para detectar ambigüedades ni para
ser socrático de forma consistente. El código Go es el socrático — determinístico,
testeable, sin alucinaciones. El LLM solo es un parser de texto → JSON.

---

Necesito crear un nuevo módulo socrático genérico que sirva como paso previo
a la creación de reglas en el Decision Engine.

El problema que resuelve:

Un usuario escribe reglas de negocio en lenguaje natural
("los perros grandes necesitan bozal"). Eso tiene que
terminar como una regla formal en Yarumo. Pero el lenguaje
natural es ambiguo: "grande" no tiene valor concreto.

Antes de que una regla entre al Decision Engine, necesito
un módulo que:

1. Reciba texto en lenguaje natural del usuario
2. Use un LLM para extraer una estructura intermedia
   (entidad, condiciones, acciones)
3. Compare esa estructura contra la ontología del dominio
   (que viene como YAML cargado en runtime, no código generado)
4. Detecte ambigüedades de forma DETERMINÍSTICA en código Go,
   no confiando en el LLM para detectarlas
5. Genere preguntas de desambiguación
6. Reciba respuestas del usuario
7. Repita hasta que no haya ambigüedades
8. Produzca una salida serializable (JSON/YAML/TOML) con la regla
   desambiguada

Las ambigüedades típicas son: valores vagos ("grande"),
campos faltantes, unidades sin referencia ("60 días"
¿desde cuándo?), valores fuera de rango.

El LLM solo parsea texto a JSON. Toda la validación y
detección de ambigüedad es código Go determinístico.

El módulo debe ser genérico: funciona con cualquier
ontología YAML, no está atado a un dominio específico.

Debe seguir la misma estructura de SDKs que ya existe
en Yarumo (sdks/decisions/core/).

## API pública

Entrada: `string` (lenguaje natural)
Salida: `[]byte` (JSON/YAML/TOML serializado)

El socrático es una caja negra desacoplada. No expone Go structs en su
frontera pública. Internamente usa Go structs, pero la API es texto → texto.
Esto permite que lo que venga antes (UI, CLI, API) y lo que venga después
(Decision Engine, otro sistema) no tengan dependencia directa con el módulo.

## Salida del módulo socrático (ejemplo conceptual)

```yaml
entity: perro
rules:
  - condition:
      type: expression  # common/expressions
      expr: "peso >= 25"
    action: requiere_bozal = true

  - condition:
      type: propositional  # logic/
      expr: "vacunas AND licencia => puede_pasear"
    action: puede_pasear = true

  - condition:
      type: predicate  # logic/predicate
      expr: "ForAll(perro, raza(perro, X) => requiere_chip(perro))"
    action: requiere_chip = true
```

El socrático usa los parsers de cada formalismo para validar que lo que
generó es sintácticamente correcto antes de emitirlo.

## Flujo

1. LLM parsea lenguaje natural → JSON con condiciones en texto
2. Go detecta qué tipo de formalismo es cada condición
3. Go parsea con el parser correspondiente para validar sintaxis
4. Go valida contra la ontología (campos, rangos, tipos)
5. Si hay ambigüedad → pregunta al usuario
6. Repite hasta limpio
7. Emite salida serializada (JSON/YAML/TOML) con expresiones formales válidas

## Arquitectura

```
sdks/socratic/core/
│
├── ontology/          # Carga y representa el dominio
│   ├── types.go       # Entity, Field, Constraint, Range, Unit
│   └── loader.go      # Loader interface (from YAML, DB, etc.)
│
├── extract/           # LLM → Go structs
│   ├── types.go       # RawIntent, RawCondition, RawAction
│   ├── client.go      # LLMClient interface (inyectado)
│   └── parser.go      # JSON response → RawIntent
│
├── analyze/           # Detección determinística de ambigüedades
│   ├── types.go       # Ambiguity (vague, missing, out_of_range, no_unit)
│   ├── analyzer.go    # Compara RawIntent vs Ontology → []Ambiguity
│   └── formalism.go   # Detecta tipo (expression/propositional/predicate)
│                      # y valida sintaxis con los 3 parsers
│
├── question/          # Genera preguntas de desambiguación
│   ├── types.go       # Question, Answer
│   └── generator.go   # []Ambiguity → []Question
│
├── resolve/           # Aplica respuestas del usuario
│   └── resolver.go    # RawIntent + []Answer → RawIntent (mejorado)
│
├── schema/            # Representación de salida
│   └── types.go       # Rule, Condition (type+expr), Action
│
└── session/           # Orquesta el loop
    ├── types.go       # Session, State, Config
    └── service.go     # El loop: extract → analyze → question → answer → repeat
```

### Flujo detallado

```
"los perros grandes necesitan bozal"
        │
        ▼
   ┌─────────┐
   │ extract/ │  LLM → JSON → RawIntent{entity:"perro", condition:"grande", action:"bozal"}
   └────┬─────┘
        ▼
   ┌──────────┐
   │ analyze/ │  RawIntent vs Ontology → Ambiguity{"grande" es vago, no hay valor}
   └────┬─────┘
        ▼
   ┌───────────┐
   │ question/ │  → "¿Qué significa 'grande'? ¿peso > X kg? ¿altura > Y cm?"
   └────┬──────┘
        ▼
     usuario responde: "peso mayor a 25 kg"
        │
        ▼
   ┌──────────┐
   │ resolve/ │  Aplica respuesta → RawIntent actualizado
   └────┬─────┘
        ▼
   ┌──────────┐
   │ analyze/ │  → 0 ambigüedades ✓
   └────┬─────┘
        ▼
   ┌─────────┐
   │ schema/ │  → Rule{condition: {type: expression, expr: "peso >= 25"}, action: ...}
   └─────────┘
```

### Dependencias

```
session/ → extract/ → (LLMClient interface)
         → analyze/ → ontology/
                     → common/expressions (parser)
                     → logic/ (parser)
                     → logic/predicate/ (parser)
         → question/
         → resolve/
         → schema/
```

`session/service.go` sería el equivalente de `evaluate/service.go` — el punto
de entrada que orquesta todo.

No hay dependencia con `sdks/decisions/core/`. Son módulos independientes.
Un módulo traductor externo conectaría la salida del socrático con
`schema.RuleSet` del Decision Engine.

## Ubicación

`sdks/decisions/socratic/` — companion module del Decision Engine SDK.

```
sdks/decisions/
  ├── core/        # evaluate, schema, adapters, explain, validate, repository
  ├── socratic/    # desambiguación NL → reglas serializadas (viene ANTES del core)
  ├── storage/     # (futuro, viene DESPUÉS del core)
  ├── endpoints/   # (futuro)
  └── ...
```

### Por qué no `modules/`
Demasiado elaborado y específico para ser un bloque reutilizable puro.
Modules son infraestructura genérica (common, math, managed).

### Por qué no `sdks/socratic/` independiente
No tiene ecosistema propio de companions. Existe para alimentar al
Decision Engine.

### Por qué `sdks/decisions/socratic/`
- Vive en el ecosistema del Decision Engine
- No depende de `core/` en código, pero sí conceptualmente
- Su salida alimenta al Decision Engine (el DaaS hace la traducción)
- Es un módulo reutilizable pero no gratis — licenciamiento propio

### Licenciamiento
El socrático es el diferenciador del producto DaaS. Es lo que permite
que un usuario no técnico cree reglas en lenguaje natural. Sin él, el
Decision Engine solo lo usan desarrolladores.

No es Apache 2.0 como el resto de Yarumo. Licencia por definir (open core,
dual license, o propietario). Es la pieza con mayor valor de mercado del
ecosistema — no se regala.

### Valor de mercado
Como componente standalone: bajo. Como feature del DaaS: es el moat.
No existe competencia directa — nadie hace desambiguación determinística
de NL contra ontologías con loop socrático en código (no LLM).
La competencia usa LLMs para todo (alucinaciones incluidas) o formularios
rígidos (sin NL).

## Pipeline completo (el socrático no funciona solo)

El socrático es texto → texto. Necesita componentes antes y después para
ser útil en un producto real.

### Antes del socrático
- **Extractor de documentos (KnowledgeForge)** — PDF/circular/ISO → texto
  plano. Genérico, no exclusivo de decisions. Podría ser su propio SDK/módulo.
- **Ontología del dominio** — definición de entidades, campos, rangos, unidades.
  Es configuración (YAML/JSON), no código. El usuario o admin del producto
  la define.
- **Interfaz de usuario** — campo de texto en UI, WhatsApp, chat de Aluna.
  Es del producto (DaaS frontend, canal), no del SDK.

### Después del socrático
- **Traductor** — salida serializada del socrático → `schema.RuleSet` para
  el Decision Engine. Vive en el producto (DaaS app), porque es quien decide
  a qué paradigma traducir.
- **Persistencia** — guardar la regla en DB. Ya existe: `repository.Repository`
  en core, implementación en la app.
- **Validación** — verificar que la regla traducida es válida. Ya existe:
  `validate/` en core.

### Quién es dueño de cada pieza

| Componente | Dueño | Por qué |
|---|---|---|
| KnowledgeForge | TBD (¿módulo propio?) | Genérico, no solo para decisions |
| Ontología | Configuración del producto | No es código |
| Interfaz de usuario | Producto (DaaS, Aluna) | Canal, no lógica |
| Socrático | `sdks/decisions/socratic/` | Companion del Decision Engine |
| Traductor | Producto (DaaS app) | Decide el paradigma destino |
| Persistencia | `core/repository/` + app | Interface en SDK, impl en app |
| Validación | `core/validate/` | Ya existe en el SDK |

## Dos ontologías (idea en exploración, no decisión)

La promesa de las ontologías es que un LLM/agente "se porte bien" (no alucine).
Pero eso no es cierto solo con un documento — se necesita combinar LLM con
código determinístico que valide contra la ontología.

Existen **dos representaciones** del mismo dominio:

### Ontología para humanos
- Documento (Word, MD) que el usuario lee
- Describe entidades, significados, reglas en lenguaje natural
- Es documentación de cara al usuario

### Ontología para el sistema
- Representación estructurada (YAML, JSON) que el código usa en runtime
- El socrático valida contra ella durante la desambiguación
- Los agentes/LLMs validan su comportamiento contra ella en runtime
- Es el **ancla de realidad** — el LLM puede alucinar, pero el código compara
  su output contra la ontología y dice "eso no existe" o "está fuera de rango"

```
Dominio real
    │
    ├──▶ Ontología humana (doc Word/MD) → el usuario la lee y entiende
    │
    └──▶ Ontología de sistema (estructurada) → el código la usa en runtime
              │
              ├── El socrático valida contra ella (desambiguación)
              └── El agente/LLM valida contra ella (no alucinar)
```

Sin la ontología de sistema, la ontología es solo un documento bonito que
el LLM puede ignorar. La clave es que **código Go determinístico** sea quien
valida, no el LLM.

Preguntas abiertas:
- ¿Quién crea la ontología de sistema? ¿El usuario? ¿Un admin? ¿Se genera
  desde la ontología humana?
- ¿Es la salida del socrático parte de la ontología de sistema, o la consume?
- ¿Cómo se mantienen sincronizadas ambas ontologías?
- ¿La ontología de sistema es input del socrático, del agente, o de ambos?

## Evolución de la ubicación (en exploración)

La ubicación del socrático ha cambiado a medida que se entiende mejor el
problema. Originalmente en `sdks/decisions/socratic/`, pero hay señales de
que la ontología es un concepto más grande que decisions.

### Validador ontológico en runtime

Cualquier sistema que use LLM (agente, chatbot, pipeline, API) necesita
validar el output del LLM contra la ontología en runtime. No es específico
de agentes ni de decisions — es infraestructura para cualquier integración
con LLM.

No es `modules/` (es valor de mercado, no se regala como Apache 2.0).

### Posible estructura: `sdks/ontology/`

```
sdks/ontology/          → SDK de ontologías (licencia propia)
  ├── socratic/         → NL → ontología de sistema + reglas desambiguadas
  ├── runtime/          → valida LLM/agentes contra ontología en runtime
  ├── ...

sdks/decisions/         → motor de reglas (consumidor de ontología)
  ├── core/             → evaluate, schema, adapters, explain, validate, repository
  ├── ...
```

Features de `sdks/ontology/`:
1. **Definición** — tipos para Entity, Field, Constraint, Range, Unit, Relationship
2. **Carga** — desde YAML/JSON/DB (interface)
3. **Validación de datos** — ¿campo existe? ¿en rango? ¿tipo correcto? ¿unidad válida?
4. **Validación de output LLM** — rechazar/marcar lo que no cuadra con la ontología
5. **Generación de constraints para prompts** — JSON Schema / instrucciones para reducir alucinaciones
6. **Detección de drift** — comparar versiones, detectar cambios breaking

### La circularidad no resuelta

Hay algo circular que no cuadra todavía:
- El socrático **produce** la ontología de sistema (NL → estructura)
- Pero el socrático **necesita** una ontología previa para desambiguar
  ("grande" es ambiguo solo si la ontología dice que "perro" tiene campo "peso"
  con rango 0-100kg)
- ¿El socrático construye la ontología incrementalmente desde cero?
- ¿O necesita una ontología semilla que después enriquece?

Hasta que esto se resuelva, la ubicación exacta del socrático y su relación
con `sdks/ontology/` queda abierta.