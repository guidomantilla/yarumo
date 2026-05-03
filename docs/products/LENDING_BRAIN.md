# LendingBrain — Motor de Decisiones para Originacion de Credito

Vertical: FinTech / Lending. Encaje con yarumo: 90%. Score: ★★★★★.

Referencia: `docs/DOMAIN_ANALYSIS.md` seccion 6a (FinTech) y seccion 7.

---

## 1) El problema

Una fintech o banco que origina credito hoy tiene:

- **Scoring ML black-box** que dice "0.73" pero no explica por que
- **Reglas de negocio en codigo hardcoded** que solo el developer entiende
- **Cero trazabilidad** — cuando el regulador pregunta "por que rechazaron a este cliente?", nadie puede reconstruir la decision
- **Calibracion a ciegas** — "que pasa si bajamos el score minimo de 650 a 600?" → nadie sabe el impacto sin probarlo en produccion
- **Proceso manual** para casos grises — analistas revisando expedientes sin criterios formales consistentes

El problema no es que no tengan ML. Es que ML predice pero no decide, no explica, y no audita. Y la SFC exige las tres cosas.

---

## 2) Que hace

LendingBrain es el motor que toma la prediccion del ML (o cualquier dato), aplica politicas de credito formales, decide, explica cada decision, y genera auditoria completa. No reemplaza el scoring ML — lo complementa con razonamiento formal.

---

## 3) Como usa cada paradigma

| Paradigma | Pregunta que responde | Ejemplo concreto |
|---|---|---|
| **Deductive** | "Cumple las politicas de credito?" | "Si edad < 18 → rechazar. Si score < 500 AND sin codeudor → rechazar. Si monto > 5x ingreso → rechazar" |
| **Bayesian** | "Cual es la probabilidad real de default?" | P(default \| score ML, ingresos, antiguedad laboral, sector economico, historial interno) — actualiza priors con data propia |
| **Fuzzy** | "Cual es el nivel de riesgo integral?" | "riesgo es ALTO cuando capacidad_pago es BAJA y estabilidad_laboral es BAJA y endeudamiento es ALTO" — no todo es binario |
| **Causal** | "Que pasa si cambio la politica?" | "Si bajo el score minimo de 650 a 600: +12% aprobaciones, +3.2% mora estimada, -$X en provisiones" |
| **MCDM** | "Cual producto ofrecerle?" | Ranking de productos dados: tasa, plazo, riesgo, rentabilidad, probabilidad de aceptacion, LTV |
| **Temporal** | "Plazos regulatorios?" | Respuesta a solicitud de credito, PQR, habeas data. "Si no respuesta EN 15 dias → incumplimiento" |
| **Predicados** | "Expediente completo?" | `PARA TODO documento EN expediente: documento.vigente = true AND documento.verificado = true` |
| **Explain/Audit** | "Por que se rechazo?" | "Rechazado por: regla R3 (endeudamiento > 60% ingreso), riesgo fuzzy ALTO (0.82), P(default) = 0.34 > umbral 0.25. Documentos verificados: 4/4. Decision tomada 2026-03-03 14:22:01" |

---

## 4) Flujo de originacion

```
1. Cliente solicita credito (app/web/sucursal)

2. Datos entran al motor:
   - Score buro (TransUnion/Datacredito/Experian)
   - Score ML interno (si existe)
   - Datos socioeconomicos (ingresos, empleo, antiguedad)
   - Historial interno (creditos previos, comportamiento de pago)
   - Expediente documental (cedula, extractos, certificados)

3. Motor evalua en cascada:

   Predicados  → Expediente completo? Documentos vigentes?
       ↓ si completo
   Deductive   → Cumple politicas minimas? (knock-out rules)
       ↓ si pasa
   Bayesian    → P(default) actualizada con priors + evidencia
       ↓
   Fuzzy       → Riesgo integral (capacidad, estabilidad, endeudamiento)
       ↓
   MCDM        → Cual producto ofrecer? (tasa, plazo, monto optimo)
       ↓
   Causal      → Sensitivity: que factor cambiaria la decision?

4. Output:
   - APROBAR / RECHAZAR / REVISAR MANUAL
   - Producto recomendado (si aprobado)
   - Explicacion completa por paradigma
   - Condiciones (monto maximo, tasa, plazo)
   - Score compuesto (combinacion de todos los paradigmas)
   - Audit trail completo
```

---

## 5) Tipos de decision que resuelve

| Decision | Hoy | Con LendingBrain |
|---|---|---|
| Aprobacion/rechazo | Score ML + reglas hardcoded | 5 paradigmas + explicacion formal |
| Monto maximo | Formula fija (30% ingreso) | Fuzzy + Bayesian (capacidad real vs nominal) |
| Tasa de interes | Segmento fijo (A/B/C/D) | MCDM (riesgo + competencia + rentabilidad + retencion) |
| Producto | El analista elige | MCDM ranking automatico con justificacion |
| Caso gris (revisar manual) | Criterio del analista (inconsistente) | Motor sugiere + explica por que es gris + que info resolveria la ambiguedad |
| Calibracion de politicas | Trial & error en produccion | Causal: simular impacto ANTES de cambiar |
| Respuesta a regulador | Buscar en logs + reconstruir manualmente | Audit trail instantaneo, trace por decision |

---

## 6) El diferenciador: la cascada explicable

Lo que hace unico a LendingBrain no es ningun paradigma individual — es la **cascada**. Cada paradigma ve el problema desde un angulo diferente y el resultado es una decision que nadie puede cuestionar:

```
"Se rechazo porque:
 1. Politicas: cumple minimas (edad OK, monto OK)
 2. Bayesian: P(default) = 0.34 (umbral: 0.25) — factores principales:
    sector economico de riesgo (0.12), antiguedad laboral < 6 meses (0.08)
 3. Fuzzy: riesgo integral = 0.82 (ALTO) — capacidad_pago BAJA (0.7),
    endeudamiento ALTO (0.85)
 4. Causal: si el cliente presentara codeudor con ingreso > $3M,
    P(default) bajaria a 0.18 → aprobaria

 Recomendacion: rechazar. O solicitar codeudor para reevaluar."
```

Eso no lo da ningun modelo ML. Eso no lo da ningun motor de reglas simple. Eso es inferential decision-making.

---

## 7) Mercado

### Quien paga

| Segmento | Tamano CO | Volumen de decisiones | Dolor |
|---|---|---|---|
| Bancos tradicionales | ~30 | Miles/dia | Regulacion SFC, fair lending, explicabilidad |
| Fintech de credito | ~80+ | Cientos-miles/dia | Escalar sin perder control, SFC sandbox |
| Cooperativas | ~180 | Decenas-cientos/dia | Mismo regimen SFC, menos tech |
| Microfinanciera | ~40+ | Cientos/dia | Volumen alto, margenes bajos, mora critica |
| Retail (credito de consumo) | Grandes cadenas | Miles/dia | Credito rotativo, tarjetas private label |

### Numeros del mercado

- Cartera de credito Colombia: ~$600B COP (2025)
- Fintech lending LATAM: crecimiento ~30% anual
- Cada punto porcentual de mora evitada = millones en provisiones ahorradas
- Cada dia de reduccion en tiempo de aprobacion = mas conversion

---

## 8) Competencia

| Competidor | Que hace | Debilidad vs LendingBrain |
|---|---|---|
| FICO Decision Manager | Scoring + reglas + ML | $$$$$, lock-in, no causal, no fuzzy |
| Mambu (decision engine) | Motor de reglas para lending | Solo reglas, no bayesian/fuzzy/causal |
| nCino | CRM + originacion | Mas CRM que motor de decisiones |
| Provenir | No-code decision engine | Reglas + ML, no explicabilidad formal, no causal |
| Interno (codigo custom) | "Lo hicimos nosotros" | Imposible de auditar, no escalable, bus factor = 1 |
| Zest AI | ML explicable | ML-centric, no multi-paradigma, pricing US |

**Diferenciador**: LendingBrain no es "otro motor de reglas" ni "otro modelo ML". Es el unico que combina 5 paradigmas con explicabilidad nativa. Cuando la SFC pregunta "por que", la respuesta es instantanea, formal, y completa.

---

## 9) Regulacion que lo hace obligatorio

| Norma | Que exige | Como LendingBrain cumple |
|---|---|---|
| Circular Externa 026/2017 SFC | Gestion de riesgo de credito (SARC) | Bayesian + Fuzzy para riesgo, Deductive para politicas, Audit para trazabilidad |
| Ley 1266/2008 (Habeas Data) | Derecho a saber por que fue reportado/rechazado | Explain/ genera justificacion completa |
| Circular Basica Contable SFC | Provisiones por calificacion de cartera | Stats (distribuciones) + Bayesian (probabilidad de deterioro) |
| Ley 1328/2009 | Proteccion al consumidor financiero | Explain/ + Temporal (plazos de respuesta) |
| SARLAFT | Anti-lavado en originacion | Deductive (reglas AML) + Bayesian (riesgo de operacion sospechosa) |
| Fair lending (tendencia global) | No discriminacion en credito | Causal: "la decision habria sido diferente si cambio genero/etnia?" — contrafactual |

La ultima fila es clave: **fair lending**. La tendencia global (EU AI Act, US ECOA) exige demostrar que las decisiones de credito no discriminan. El contrafactual causal ("habria sido diferente si...?") es exactamente lo que los reguladores empiezan a exigir. Esto no lo resuelve ML — lo resuelve razonamiento causal.

---

## 10) Stack tecnico

```
┌─────────────────────────────────────────────┐
│  Frontend (portal analista + portal cliente) │
├─────────────────────────────────────────────┤
│  API REST/gRPC                               │
│  ┌──────────────────────────────────────┐    │
│  │ Orchestrator (cascada de decision)   │    │
│  │ Pre-check → KnockOut → Score →      │    │
│  │ Risk → Product → Sensitivity        │    │
│  └──────────────────────────────────────┘    │
├─────────────────────────────────────────────┤
│  SDK decision-engine/core/                   │
│  (decisions + explain + validate + audit)    │
├─────────────────────────────────────────────┤
│  yarumo compute/ + maths/                  │
│  (deductive, bayesian, fuzzy, causal, mcdm) │
├─────────────────────────────────────────────┤
│  Integraciones externas                      │
│  ┌────────────┬───────────┬──────────────┐  │
│  │ Buro       │ Core      │ Doc          │  │
│  │(Datacredito│ bancario  │ verification │  │
│  │  Experian) │ (APIs)    │ (OCR/LLM)   │  │
│  └────────────┴───────────┴──────────────┘  │
├─────────────────────────────────────────────┤
│  Storage + Telemetry + NLP companion         │
└─────────────────────────────────────────────┘
```

---

## 11) Go-to-market

**Fase 1**: Una fintech de credito como design partner. Caso: microcredito o credito de libre inversion. Implementar politicas de knock-out + scoring bayesian + explicabilidad. Demostrar reduccion en tiempo de aprobacion y generacion automatica de explicacion de rechazo.

**Fase 2**: Agregar fuzzy risk scoring + MCDM product recommendation. Demostrar que el motor selecciona mejor producto que el analista humano.

**Fase 3**: Agregar causal "what-if" para calibracion de politicas. Vender al comite de credito: "antes de cambiar la politica, simulen el impacto".

**Fase 4**: Bancos medianos — el argumento es regulatorio: "la SFC va a exigir explicabilidad formal. Preparese antes, no despues de la multa."

---

## 12) Por que 90% y no 100%

| Gap | Que es | Como se resuelve |
|---|---|---|
| Score ML | El modelo predictivo (XGBoost, neural net) | Externo — LendingBrain consume el score, no lo genera |
| Integracion buro | Datacredito, TransUnion, Experian APIs | Infra, no inferencia |
| OCR / verificacion documental | Extraer datos de cedula, extractos | LLM/OCR companion, no inferencia |
| Core bancario | Desembolso, contabilizacion | Integracion, no inferencia |

El motor de inferencia hace el 90% del trabajo intelectual de la decision. El 10% es plomeria de integracion.
