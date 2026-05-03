# KnowledgeForge — Pipeline de Extraccion de Conocimiento Normativo

El gap comun a todos los productos: de donde vienen las reglas?

Referencia: `docs/DOMAIN_ANALYSIS.md`, `docs/AGENTIC_PRODUCTS.md`, `docs/DECISION_ENGINE_ARCH.md` (nlp/ companion module).

---

## 1) El problema

Todos los productos (ComplianceEngine, LendingBrain, UnderwriteIQ, GovRules, ClinicalRules, AgroDecide) asumen que las reglas ya estan en el motor. Pero en la realidad:

```
Circular SFC  ──→ PDF ──→ ???  ──→ Reglas formales
ISO 9001      ──→ PDF ──→ ???  ──→ Reglas formales
GPC MinSalud  ──→ PDF ──→ ???  ──→ Reglas formales
Guia Cenicafe ──→ PDF ──→ ???  ──→ Reglas formales
Manual interno──→ Word ──→ ??? ──→ Reglas formales

                          ↑
                    ESTE GAP
                    es donde
                    todo se cae
```

Hoy ese `???` es un humano experto que lee el documento, interpreta, y escribe reglas a mano. Eso es:

- **Lento** — semanas por documento
- **Caro** — necesitas experto de dominio + ingeniero
- **Inconsistente** — dos personas extraen reglas diferentes del mismo doc
- **No escala** — una empresa tiene cientos de documentos normativos

---

## 2) El angulo ISO: conocimiento propietario

Cada empresa tiene ISOs. Cada implementacion de ISO es UNICA:

| ISO | Dominio | Que contiene | Por que es unico por empresa |
|---|---|---|---|
| 9001 | Calidad | Procesos, controles, criterios de aceptacion | Cada planta tiene sus procesos, sus tolerancias, sus puntos de control |
| 14001 | Ambiental | Aspectos ambientales, controles, limites | Cada planta tiene sus emisiones, sus residuos, sus riesgos |
| 45001 | Seguridad | Peligros, controles, EPP, procedimientos | Cada planta tiene sus riesgos, sus maquinas, sus zonas |
| 27001 | InfoSec | Controles de seguridad, politicas, activos | Cada empresa tiene su arquitectura, sus datos, sus riesgos |
| 22000 | Alimentos | HACCP, puntos criticos, limites | Cada planta tiene sus procesos, sus materias primas, sus PCC |
| 50001 | Energia | Usos significativos, lineas base, metas | Cada planta tiene sus equipos, sus consumos, sus objetivos |

Los ISOs tienen estructura que es oro para extraccion automatica:

- **Lenguaje normativo**: "shall" / "must" / "should" → reglas deductivas
- **Criterios de aceptacion**: "dentro de rango X-Y" → reglas fuzzy / deductivas
- **Frecuencias**: "cada N dias" / "antes de X" → reglas temporales
- **Matrices de riesgo**: probabilidad x impacto → bayesian + MCDM
- **No conformidades**: si criterio no se cumple → consecuencia → reglas deductivas

Lo mas importante: los documentos ISO de cada empresa son **conocimiento propietario**. No es generico — es SU implementacion, SUS procedimientos, SUS umbrales. Es el conocimiento que la empresa tardo anos en construir.

---

## 3) Que hace KnowledgeForge

Pipeline completo de extraccion de conocimiento normativo: documento → reglas formales, con humano en el loop para validacion.

### Paso 1: Ingestion

```
Input: PDF, Word, HTML, texto estructurado
- ISOs (9001, 14001, 45001, 27001, 22000, 50001...)
- Regulaciones (circulares, leyes, resoluciones)
- Guias clinicas (GPC)
- Manuales internos de procedimiento
- Politicas de credito, suscripcion, compliance
- Fichas tecnicas, instructivos de trabajo
```

### Paso 2: Extraction (LLM)

```
LLM analiza el documento y extrae:
- OBLIGACIONES: "La empresa DEBE..." → regla deductiva
- CONDICIONES: "Cuando X AND Y → Z" → regla deductiva
- UMBRALES: "Temperatura entre 18-25°C" → regla fuzzy / deductiva
- PLAZOS: "Dentro de 15 dias habiles" → regla temporal
- CRITERIOS: "Evaluar segun impacto y probabilidad" → MCDM
- PROBABILIDADES: "Si historial de incidentes..." → bayesian prior
- VERIFICACIONES: "Para todo lote: verificar X" → predicado
```

### Paso 3: Structured intermediate (human-reviewable)

```yaml
source: "ISO 9001:2015 - Seccion 8.5.1 - Control de produccion"
rules:
  - id: ISO9001-851-01
    type: deductive
    condition: "temperatura_horno > 220 OR temperatura_horno < 180"
    conclusion: "detener_linea = true, alerta_calidad = true"
    confidence: 0.95
    original_text: "La temperatura del horno debe mantenerse entre 180 y 220..."

  - id: ISO9001-851-02
    type: temporal
    condition: "calibracion_instrumento"
    constraint: "CADA 30 dias"
    consequence: "si vencido → no_conformidad_menor"
    original_text: "Los instrumentos de medicion deben calibrarse mensualmente..."

  - id: ISO9001-851-03
    type: fuzzy
    variable: "calidad_producto"
    terms:
      - name: "aceptable"
        range: [95, 100]
      - name: "marginal"
        range: [90, 95]
      - name: "rechazable"
        range: [0, 90]
    original_text: "El producto se clasifica segun el indice de calidad..."
```

### Paso 4: Human review

```
El experto de dominio revisa cada regla extraida:
- La interpretacion es correcta?
- Los umbrales son los de MI planta?
- Falta contexto que el LLM no capturo?
- Hay reglas implicitas que el documento no dice pero el experto sabe?

El experto EDITA, no escribe desde cero. 10x mas rapido.
```

### Paso 5: Deploy

```
Reglas aprobadas → inference engine
- Deductive rules → deductive/
- Bayesian priors → bayesian/
- Fuzzy variables → fuzzy/
- Temporal constraints → temporal/
- MCDM criteria → mcdm/
- Predicates → predicate/
```

---

## 4) Por que es un PRODUCTO, no solo un feature

| Como feature (NLP companion) | Como producto (KnowledgeForge) |
|---|---|
| "Extraiga reglas de este PDF" | Pipeline completo de gestion de conocimiento normativo |
| One-shot | Continuo: monitorea cambios en documentos fuente |
| Sin versionamiento | Versiona reglas: quien cambio que, cuando, por que |
| Sin validacion cruzada | Detecta contradicciones entre documentos fuente |
| Sin coverage analysis | "Su ISO 9001 tiene 47 clausulas → 42 tienen reglas, 5 gaps" |

---

## 5) Ejemplo completo: planta de alimentos

```
EMPRESA: Planta de alimentos, certificada ISO 22000 + ISO 14001

KnowledgeForge ingiere:
  - Manual HACCP (ISO 22000) → 23 puntos criticos de control
  - Manual ambiental (ISO 14001) → 15 aspectos ambientales significativos
  - Procedimientos operativos → 89 instrucciones de trabajo
  - Matriz de riesgos → 45 riesgos con probabilidad x impacto

KnowledgeForge extrae:
  - 156 reglas deductivas (shall/must del ISO)
  - 23 reglas temporales (frecuencias de monitoreo, calibracion)
  - 12 variables fuzzy (calidad, frescura, contaminacion)
  - 45 evaluaciones MCDM (matriz de riesgos)
  - 8 redes bayesianas (probabilidad de contaminacion dado factores)

Experto de calidad revisa en 2 dias (vs 2 meses manual).

Motor de inferencia ahora TIENE el conocimiento de la planta.
Puede:
  - Monitorear PCC en tiempo real (IoT + reglas)
  - Alertar cuando un control esta fuera de rango
  - Preparar la auditoria ISO automaticamente (audit trail)
  - Responder "por que se detuvo la linea?" con explicacion formal
  - Simular "que pasa si cambiamos el PCC 3?" (causal)
```

---

## 6) Tipos de documentos por vertical

| Vertical | Documentos fuente | Tipo de reglas que se extraen |
|---|---|---|
| **RegTech** | Circulares SFC, leyes, resoluciones, SARLAFT | Deductive + Temporal (plazos legales) |
| **FinTech** | Politicas de credito, manuales SARC, circulares | Deductive + Bayesian (umbrales de riesgo) |
| **InsurTech** | Condiciones generales, manuales de suscripcion, tarifarios | Deductive + Fuzzy + Stats (tablas actuariales) |
| **GovTech** | Resoluciones, CONPES, manuales operativos de programas | Deductive + Temporal + Predicados |
| **HealthTech** | GPC MinSalud, protocolos clinicos, vademecum | Deductive + Bayesian + Temporal |
| **AgriTech** | Guias Cenicafe/ICA, fichas tecnicas, protocolos BPA | Deductive + Bayesian + Fuzzy |
| **Manufactura** | ISOs (9001, 14001, 45001, 22000), instructivos de trabajo | Todo — ISOs cubren todos los paradigmas |
| **InfoSec** | ISO 27001, politicas de seguridad, matrices de riesgo | Deductive + MCDM (riesgo) + Temporal (revisiones) |

---

## 7) Capacidades avanzadas

### Coverage analysis

```
KnowledgeForge analiza:
  "Su ISO 9001:2015 tiene 51 clausulas con 'shall' (requisitos obligatorios).
   Estado actual:
   - 42 clausulas tienen reglas formales implementadas (82%)
   - 5 clausulas tienen reglas parciales (10%)
   - 4 clausulas no tienen reglas (8%)

   Gaps criticos:
   - Clausula 8.5.4 (preservacion) — sin reglas de control de preservacion
   - Clausula 9.1.3 (analisis y evaluacion) — sin reglas de tendencia
   - Clausula 10.2 (no conformidad) — reglas incompletas de accion correctiva

   Riesgo de auditoria: MEDIO — los 4 gaps son auditables por certificadora."
```

### Contradiction detection

```
KnowledgeForge detecta:
  "Contradiccion encontrada:
   - ISO 22000 PCC-7 dice: 'temperatura maxima 4°C para almacenamiento'
   - Procedimiento interno PR-ALM-03 dice: 'temperatura maxima 6°C'
   - Ficha tecnica proveedor dice: 'mantener bajo 5°C'

   Recomendacion: unificar a 4°C (el mas restrictivo, ISO prevalece).
   Impacto: 2 reglas del motor deben actualizarse."
```

### Change monitoring

```
KnowledgeForge monitorea:
  "Nueva version del procedimiento PR-CAL-01 detectada (v3.2 → v3.3).
   Cambios identificados:
   - Frecuencia de calibracion termometros: mensual → quincenal
   - Nuevo instrumento agregado: pH-metro linea 4
   - Criterio de aceptacion dureza: 45-55 HRC → 47-53 HRC (mas restrictivo)

   Reglas afectadas: 3
   Propuesta de actualizacion adjunta. Aprueba?"
```

### Version history

```
Regla ISO9001-851-01:
  v1.0 (2026-01-15) — Creada por KnowledgeForge, aprobada por J. Martinez
  v1.1 (2026-03-03) — Umbral ajustado de 220 a 215°C por hallazgo de auditoria
  v1.2 (2026-03-10) — Agregada condicion: "AND linea_activa = true" por falso positivo

  Documento fuente: ISO 9001:2015 Sec 8.5.1 + Procedimiento PR-PRD-07 v2.1
  Ultima auditoria: 2026-02-28 — conforme
```

---

## 8) Donde vive en la arquitectura

```
┌─────────────────────────────────────────────┐
│  KnowledgeForge (producto / companion)       │
│  ┌──────────┬───────────┬────────────────┐  │
│  │ Ingestion│ Extraction│ Review UI      │  │
│  │ (docs)   │ (LLM)    │ (human-in-loop)│  │
│  └──────────┴───────────┴────────────────┘  │
│  ┌──────────┬───────────┬────────────────┐  │
│  │ Validate │ Version   │ Coverage       │  │
│  │ (consist)│ (history) │ (gap analysis) │  │
│  └──────────┴───────────┴────────────────┘  │
├─────────────────────────────────────────────┤
│  SDK decision-engine/core/                   │
│  (decisions + explain + validate + audit)    │
├─────────────────────────────────────────────┤
│  yarumo compute/ + maths/                  │
│  (deductive, bayesian, fuzzy, causal, mcdm) │
└─────────────────────────────────────────────┘
```

Conecta con el `nlp/` companion module planificado en `DECISION_ENGINE_ARCH.md`, pero es mas grande que NLP — es gestion de conocimiento normativo completo.

---

## 9) El moat

El conocimiento propietario extraido es el moat. Una vez que la empresa tiene sus ISOs, sus politicas, sus manuales formalizados como reglas en el motor — cambiar de proveedor significa re-extraer todo. Y cada dia que el sistema opera, el conocimiento se refina (el experto corrige, agrega, ajusta).

El flywheel:

```
Mas documentos → mas reglas → mas valor → mas dependencia → mas documentos
```

---

## 10) Mercado especifico: ISO compliance

### Quien paga

| Segmento | Tamano CO | ISOs tipicos | Dolor |
|---|---|---|---|
| Manufactura | ~10K empresas certificadas | 9001, 14001, 45001 | Preparar auditoria = semanas de panico |
| Alimentos | ~3K plantas | 22000, HACCP, BPM | PCC, trazabilidad, INVIMA |
| Farmaceutica | ~200 laboratorios | 9001, BPM, GMP | INVIMA, FDA si exporta |
| TI / Software | ~500 empresas | 27001, 20000 | Seguridad de informacion, SOC2 |
| Energia | ~200 empresas | 50001, 14001 | CREG, eficiencia energetica |
| Construccion | ~1K empresas | 9001, 14001, 45001 | Licitaciones exigen ISO |
| Exportadores | ~5K empresas | Varios + certificaciones de mercado destino | GlobalGAP, FSSC 22000, BRC |

### Numeros

- Empresas certificadas ISO en Colombia: ~15,000+
- Costo de una auditoria de certificacion: $5-20M COP
- Costo de preparar la auditoria (interno): $20-50M COP en horas-hombre
- Costo de perder la certificacion: inmedible (pierde contratos, pierde mercados)
- KnowledgeForge reduce preparacion de auditoria de semanas a dias

### El pitch

"Sus ISOs son su conocimiento propietario. Hoy viven en PDFs que solo 3 personas entienden. KnowledgeForge los convierte en reglas ejecutables, monitoreables, y auditables. Cuando llegue el auditor, el sistema le muestra exactamente que controles existen, como se monitorean, y la evidencia de cumplimiento — automaticamente."

---

## 11) Relacion con otros productos

KnowledgeForge es el **habilitador** de todos los demas productos:

| Producto | Sin KnowledgeForge | Con KnowledgeForge |
|---|---|---|
| ComplianceEngine | Reglas escritas a mano | Reglas extraidas de circulares automaticamente |
| LendingBrain | Politicas codificadas por ingeniero | Politicas extraidas de manuales SARC |
| UnderwriteIQ | Reglas del actuario manualmente | Reglas extraidas de manuales de suscripcion |
| GovRules | Reglas de resoluciones a mano | Reglas extraidas de resoluciones/CONPES |
| ClinicalRules | GPC formalizadas por medico+ingeniero | GPC extraidas automaticamente del PDF |
| AgroDecide | Guias de Cenicafe a mano | Guias extraidas del PDF + conocimiento del extensionista |
| AgentGuard | Politicas de gobernanza escritas a mano | Politicas extraidas de frameworks de governance |

KnowledgeForge no es un producto vertical — es la **pieza horizontal que alimenta a todos los demas**. Sin el, cada producto tiene el mismo cuello de botella: alguien tiene que escribir las reglas a mano.
