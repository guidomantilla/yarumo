# Productos Agentic AI — Inference Engine + Agentes Autonomos

Los 6 productos anteriores (ComplianceEngine, LendingBrain, UnderwriteIQ, GovRules, ClinicalRules, AgroDecide) son **motores**: el humano los opera. Con Agentic AI, se convierten en **agentes** que actuan autonomamente dentro de guardrails formales.

Referencia: `docs/DOMAIN_ANALYSIS.md` (analisis de dominio), `docs/COMPLIANCE_ENGINE.md`, `docs/LENDING_BRAIN.md`, etc.

---

## La salsa especial: triple rol del inference engine

El inference engine juega 3 roles en un sistema agentico:

1. **El cerebro** — el agente usa inferencia para decidir
2. **El guardrail** — el agente esta RESTRINGIDO por reglas formales
3. **El auditor** — cada accion del agente se registra, explica, y audita

Ningun framework agentico actual (LangChain, CrewAI, AutoGen) tiene esto. Todos usan prompts para guardrails — fragiles, no auditables, no formales.

```
Agente AI quiere ejecutar accion X
        ↓
   Inference Engine evalua:
        ↓
   Deductive  → La politica permite esta accion en este contexto?
   Bayesian   → Cual es la probabilidad de consecuencia negativa?
   Fuzzy      → Que tan riesgosa es? (no todo es blanco/negro)
   Causal     → Que efectos cascada tendria?
   Temporal   → Es el momento correcto? Viola algun SLA?
        ↓
   Decision: PERMITIR / DENEGAR / ESCALAR A HUMANO
        ↓
   Explain  → Por que (para el humano supervisor)
   Audit    → Registro formal (para compliance)
```

---

## 1) AgentGuard — Governance Platform para AI Agents

### Score: ★★★★★ — el producto mas importante estrategicamente

### El problema que NADIE ha resuelto bien

Cada empresa que despliega agentes AI tiene el mismo miedo: "que pasa si el agente hace algo que no debia?" Hoy, los guardrails son:

- **Prompts** — "No hagas X" en el system prompt. Fragil. Jailbreakeable. No auditable.
- **Output filters** — Regex/ML sobre la respuesta. Reactivo, no preventivo.
- **Rate limits** — "Maximo N acciones por minuto." Burdo. No entiende contexto.
- **Human-in-the-loop siempre** — Mata la autonomia. No escala.

Ninguno responde: "PUEDE este agente hacer esta accion, en ESTE contexto, con ESTAS consecuencias, y QUEDA REGISTRO formal de por que?"

### Que hace

Plataforma que gobierna agentes AI con razonamiento formal. Define que pueden hacer (deductive), evalua riesgo de cada accion (bayesian), maneja zonas grises (fuzzy), simula consecuencias (causal), prioriza entre acciones conflictivas (mcdm), verifica plazos y secuencias (temporal), y audita todo.

### Ejemplo concreto: agente de customer service

```
Agente quiere: ofrecer descuento de 30% al cliente

AgentGuard evalua:
  Deductive:  politica permite hasta 20% sin aprobacion → EXCEDE
  Bayesian:   P(retencion cliente | 30% descuento) = 0.85
              P(retencion cliente | 20% descuento) = 0.72
  Fuzzy:      valor del cliente: ALTO (compras frecuentes, antiguedad 3 anos)
  Causal:     si aprueba 30% → precedente para otros clientes similares
              impacto estimado en margen: -$12M/ano si se generaliza
  MCDM:       20% descuento + upgrade de plan (rank 1) vs 30% descuento (rank 2)

Decision: DENEGAR 30%. SUGERIR alternativa: 20% + upgrade.
Razon: excede politica, alternativa tiene mejor balance retencion/margen.
Si el agente insiste → ESCALAR a supervisor humano.
```

### Por que ★★★★★

| Factor | Score |
|---|---|
| Encaje con yarumo | 95% — es literalmente para lo que se diseno |
| Mercado | Explosivo — toda empresa con agentes AI necesita governance |
| Timing | Perfecto — AI governance es el compliance del 2026-2030 |
| Competencia | Casi inexistente con razonamiento formal |
| Regulacion | EU AI Act, executive orders, regulaciones nacionales — OBLIGAN governance |
| Barrera de entrada | Baja — no necesita datos de dominio, el cliente define las reglas |

### Competencia

| Competidor | Que hace | Debilidad |
|---|---|---|
| Guardrails AI | Output validation basada en LLM | LLM-based → no determinista, no auditable, jailbreakeable |
| Lakera | Prompt injection detection | Solo seguridad, no governance de decisiones |
| Arthur AI | ML model monitoring | Monitoreo post-hoc, no prevencion en tiempo real |
| Custom prompts | "No hagas X" | Fragil, no formal, no auditable |

**Diferenciador brutal**: AgentGuard es el UNICO que usa razonamiento formal (determinista, completo, auditable) para gobernar agentes. No es "otro LLM vigilando al LLM" — es logica formal verificable.

---

## 2) AutoCompliance — Agente Autonomo de Cumplimiento Regulatorio

### ComplianceEngine era un motor. AutoCompliance es un agente.

| ComplianceEngine (motor) | AutoCompliance (agente) |
|---|---|
| El oficial de cumplimiento define reglas | El agente LEE la circular/ley y propone las reglas |
| El oficial ejecuta evaluaciones | El agente monitorea y evalua proactivamente |
| El oficial genera reportes | El agente genera Y ENVIA reportes al regulador |
| El oficial responde PQRs | El agente redacta la respuesta fundamentada |
| Reactivo: "evalua esto" | Proactivo: "detecte esto, ya actue, mire el reporte" |

### Flujo autonomo

```
1. MONITOREO (continuo, autonomo):
   - LLM lee Diario Oficial, circulares SFC, resoluciones Supersalud
   - LLM extrae reglas nuevas → traduce a formato formal
   - Inference engine compara con reglas actuales → identifica GAPS
   - Agente notifica: "Nueva circular 042/2026 de SFC. Impacto: 3 reglas
     nuevas de SARLAFT. Gap vs su ruleset actual: 2 reglas faltantes.
     Propuesta de reglas adjunta. Aprueba?"

2. EVALUACION (diaria, autonoma):
   - Agente cruza bases de datos (clientes, transacciones, plazos)
   - Inference engine evalua cumplimiento por entidad/proceso
   - Agente genera dashboard de estado
   - Si detecta incumplimiento → alerta inmediata con explicacion

3. RESPUESTA (por evento, semi-autonoma):
   - Llega PQR / derecho de peticion / requerimiento regulatorio
   - Agente analiza el requerimiento (LLM)
   - Inference engine evalua los hechos vs las reglas
   - Agente redacta respuesta fundamentada (LLM + Explain/)
   - Humano revisa y aprueba → agente envia
   - Temporal: monitorea deadline, escala si se acerca al vencimiento

4. REPORTE (periodico, autonomo):
   - Agente genera reportes regulatorios (SARLAFT, ROS, indicadores)
   - Inference engine verifica completitud y consistencia
   - Agente formatea segun requerimiento del regulador
   - Humano aprueba → agente transmite
```

### El humano trabaja 10x menos

```
HOY (sin agente):
  Oficial lee circular → interpreta → escribe reglas → programa en sistema →
  ejecuta evaluacion → analiza resultados → escribe reporte → envia
  Tiempo: 2-4 semanas por circular nueva

CON AutoCompliance:
  Agente lee circular → propone reglas → [HUMANO APRUEBA] →
  agente evalua automaticamente → genera reporte → [HUMANO APRUEBA] → envia
  Tiempo: 1-2 dias (90% es esperar aprobacion humana)
```

---

## 3) CreditAgent — Agente Autonomo de Originacion de Credito

### LendingBrain era un motor. CreditAgent maneja la originacion completa.

### Flujo conversacional

```
1. CONVERSACION (agente ↔ solicitante via WhatsApp/chat):

   Agente: "Hola Maria, vi que inicio su solicitud de credito.
           Necesito verificar algunos datos. Su ingreso mensual actual?"
   Maria:  "Tres millones doscientos"
   Agente: [LLM extrae: ingreso = $3.2M COP]
           [Deductive: monto solicitado $15M, ratio 4.7x ingreso — dentro de politica]
           "Perfecto. Puede enviarme foto de su ultimo extracto bancario?"
   Maria:  [envia foto]
   Agente: [OCR/LLM extrae datos del extracto]
           [Bayesian: actualiza P(default) con datos reales vs declarados]
           [Predicados: verifica extracto vs ingreso declarado — consistente]
           "Gracias. Su solicitud esta pre-aprobada por $12M a 36 meses.
            Le explico por que $12M y no $15M:"
           [Explain/: "Monto maximo = 3.75x ingreso verificado.
            Ingreso verificado: $3.2M. Maximo: $12M. Regla: politica R-004."]

2. DECISION (autonoma dentro de politicas):

   [Cascada completa del inference engine]
   Si APROBAR y monto < $20M → agente aprueba directamente
   Si APROBAR y monto > $20M → escala a analista humano
   Si RECHAZAR → agente explica y sugiere alternativas
   Si CASO GRIS → agente reune mas info antes de escalar

3. POST-DECISION (agente maneja seguimiento):

   Agente: "Maria, su credito fue aprobado. Le envio los documentos
           para firma electronica. Tiene alguna pregunta sobre
           las condiciones?"
   Maria:  "Por que la tasa es 1.8% y no menos?"
   Agente: [Explain/: "Su tasa se calculo asi: tasa base 1.2% +
           ajuste por score de riesgo MEDIO (+0.4%) + ajuste por
           plazo 36 meses (+0.2%). Para obtener mejor tasa:
           presente un codeudor (estimado: -0.3%) o reduzca
           plazo a 24 meses (-0.15%)."]
```

### Motor vs Agente

| LendingBrain (motor) | CreditAgent (agente) |
|---|---|
| Analista humano recoge datos | Agente conversa y recoge datos |
| Analista alimenta el motor | Agente se auto-alimenta |
| Motor decide, analista comunica | Agente decide Y comunica |
| Cliente espera 2-5 dias | Cliente tiene respuesta en 15 minutos |
| Analista maneja 20 solicitudes/dia | Agente maneja 2,000/dia |
| Explicacion disponible si la piden | Explicacion proactiva en cada paso |

---

## 4) ClaimsAgent — Agente Autonomo de Gestion de Siniestros

### Producto nuevo. No es UnderwriteIQ (suscripcion) — es la otra mitad del seguro.

### Flujo

```
1. AVISO DE SINIESTRO (agente ↔ asegurado):

   Asegurado: "Me chocaron el carro en la Calle 80 con Boyaca"
   Agente: [LLM entiende: siniestro auto, choque, ubicacion]
           "Lamento escuchar eso. Esta usted bien? Hay heridos?"
   Asegurado: "No, solo danos materiales"
   Agente: [Deductive: siniestro sin lesiones → protocolo danos materiales]
           "Necesito que me envie:
            1. Fotos del dano (minimo 4 angulos)
            2. Foto del otro vehiculo
            3. Foto del parte policial o acuerdo entre partes"

2. ANALISIS (autonomo):

   [LLM/CV analiza fotos → estima dano]
   [Deductive: cobertura vigente? deducible? exclusiones?]
   [Bayesian: P(fraude) = 0.03 — siniestro consistente, cliente sin historial]
   [Fuzzy: severidad MEDIA — dano visible pero no estructural]
   [MCDM: taller preferido mas cercano vs taller del cliente]
   [Temporal: respuesta al asegurado en < 24 horas (poliza)]

3. DECISION (autonoma para siniestros simples):

   Si P(fraude) < 0.1 AND monto estimado < $5M AND cobertura clara:
     → Agente APRUEBA directamente
     → Genera orden de reparacion
     → Notifica al asegurado con explicacion completa

   Si P(fraude) > 0.3 OR monto > $20M OR cobertura ambigua:
     → ESCALA a ajustador humano con analisis completo
     → "Este caso requiere revision manual.
        Razon: monto estimado $22M excede umbral de aprobacion autonoma.
        Analisis adjunto: cobertura OK, fraude bajo, pero verificar
        peritaje presencial del dano estructural."

4. OUTPUT:

   "Estimado Juan, su siniestro #2026-0342 ha sido APROBADO.
    Resumen:
    - Dano: impacto lateral derecho, guardafango + puerta
    - Monto aprobado: $3.8M COP
    - Deducible: $500K (usted paga)
    - Taller asignado: AutoCenter Calle 80 (2km de su ubicacion)
    - Cita disponible: manana 8am o pasado manana 2pm
    Cual prefiere?

    Detalle de la decision:
    - Poliza 2025-AUTO-1234 vigente hasta 2026-08-15 ✓
    - Cobertura danos a terceros + propios ✓
    - Sin exclusiones aplicables ✓
    - P(fraude): 0.03 (bajo) — siniestro consistente con fotos y ubicacion"
```

### Numeros

- Costo promedio de gestion de un siniestro auto: ~$200-500K COP (ajustador + administrativo)
- Volumen Colombia: ~2M siniestros auto/ano
- El 70% de siniestros auto son "simples" (sin lesiones, monto bajo)
- Si ClaimsAgent maneja el 70% simple autonomamente → ahorro masivo + respuesta en horas vs dias

---

## 5) PolicyForge — Agente Disenador de Politicas con Simulacion Causal

### Producto completamente nuevo. No existe equivalente.

El problema: disenar politicas (de credito, seguros, compliance, gobierno) hoy es trial-and-error. Cambias un umbral, lo pones en produccion, y esperas 3 meses para ver el impacto.

PolicyForge es un agente que DISENA politicas usando simulacion causal:

```
Humano: "Quiero reducir la mora del portafolio de microcredito
         sin bajar la aprobacion mas de 10%"

PolicyForge:
  1. [Lee ruleset actual completo]
  2. [Analiza datos historicos: quienes entraron en mora y por que]
  3. [Causal: identifica los 3 factores que mas causan mora]
     "Los factores causales principales son:
      - Endeudamiento > 55% del ingreso (contribucion: 34%)
      - Antiguedad laboral < 4 meses (contribucion: 28%)
      - Sector economico informal sin aval (contribucion: 19%)"

  4. [Simula 50 variantes de politica modificando umbrales]
  5. [MCDM: rankea variantes por: mora estimada, aprobacion, rentabilidad]

  6. "Recomiendo la variante 17:
      - Bajar umbral de endeudamiento de 60% a 52%
      - Agregar regla: antiguedad < 4 meses → requiere codeudor
      - Mantener sector informal sin cambios (impacto bajo vs costo social)

      Impacto simulado:
      - Mora estimada: 4.2% → 3.1% (-1.1pp)
      - Aprobacion: 68% → 61% (-7pp, dentro del -10% permitido)
      - Rentabilidad: +$340M/ano en provisiones ahorradas

      Quiere ver las otras 49 variantes? Ajustar restricciones?
      Aprobar para implementacion?"

  7. [Si aprueba → genera las reglas formales → validate/ verifica
      contradicciones y gaps → despliega en el motor]
```

### Por que no existe

Porque requiere CAUSAL inference (no solo correlacion) + motor de reglas formal + simulacion + explicabilidad. Cada pieza existe por separado, pero nadie las ha integrado con un agente que las orqueste.

---

## 6) AuditBot — Agente Autonomo de Auditoria de Decisiones

### Audita las decisiones de OTROS sistemas/agentes.

```
AuditBot recibe: log de 10,000 decisiones de credito del ultimo mes

AuditBot analiza autonomamente:
  1. [Deductive: cada decision siguio las reglas vigentes?]
     → 47 decisiones violaron regla R-012 (aprobaron sin codeudor
        cuando endeudamiento > 55%)

  2. [Bayesian: hay patrones anomalos?]
     → Analista #7 aprueba 23% mas que el promedio.
        P(sesgo) = 0.82.
        Especificamente en montos $10-15M, aprueba 40% mas.

  3. [Causal: las violaciones causaron dano?]
     → De las 47 violaciones de R-012: 12 ya estan en mora (25.5%)
        vs 4.2% del portafolio general. Impacto estimado: $89M en provisiones.

  4. [MCDM: que hallazgos priorizar?]
     → Ranking: (1) Sesgo analista #7 — urgente, (2) Violaciones R-012 —
        politica debe reforzarse, (3) 3 aprobaciones sin documentos completos.

  5. [Temporal: plazos de reporte?]
     → Reporte a comite de credito: vence viernes.
        Reporte SARLAFT mensual: vence dia 10.

Reporte generado automaticamente. Enviar a comite?
```

Aplica a TODO: audita decisiones de credito, suscripcion, claims, gobierno, compliance, incluso decisiones de OTROS agentes AI. AuditBot es el auditor que nunca duerme.

---

## 7) WhatsApp Advisor — Agente de Dominio via Mensajeria

### El pattern que conecta con LATAM

En LATAM, WhatsApp es la plataforma. No apps, no portales web — WhatsApp. Este no es un producto, es un **channel pattern** que aplica a todos los productos:

```
CAFETERO en Planadas (AgroDecide):
  "Don Julio, buenas tardes. Las condiciones de humedad en su zona
   llevan 5 dias sobre 85%. El lote 3 (Caturra) tiene riesgo alto
   de roya. Le recomiendo aplicar fungicida esta semana.
   Responda SI para ver el detalle o AYUDA para hablar con su
   extensionista."

SOLICITANTE de credito (CreditAgent):
  "Maria, su solicitud esta pre-aprobada por $12M.
   Le envio las condiciones. Tiene preguntas?"

ASEGURADO (ClaimsAgent):
  "Juan, su siniestro fue aprobado. Cita en AutoCenter
   manana 8am. Responda 1 para confirmar, 2 para otro horario."

BENEFICIARIO de subsidio (GovRules):
  "Senora Garcia, su solicitud a Colombia Mayor fue aprobada.
   Recibira el primer pago el 15 de abril.
   Si tiene preguntas, responda AYUDA."
```

El agente RAZONA con el inference engine pero HABLA por WhatsApp. La formalidad esta adentro, la simplicidad esta afuera.

---

## 8) Mapa completo: productos con Agentic AI

```
                    yarumo (inference engine)
                            │
              ┌─────────────┼──────────────┐
              │             │              │
         GOVERNANCE    VERTICAL       AUTONOMOUS
              │        AGENTS          AGENTS
              │             │              │
         AgentGuard    CreditAgent    AutoCompliance
         (horizontal)  ClaimsAgent    AuditBot
                       PolicyForge
                            │
                    WhatsApp Advisor
                    (channel pattern)
```

### 3 categorias

| Categoria | Producto | Que hace | El engine es... |
|---|---|---|---|
| **Governance** | AgentGuard | Gobierna cualquier agente AI | El guardrail + auditor |
| **Vertical Agents** | CreditAgent, ClaimsAgent, PolicyForge | Agentes especializados que hacen el trabajo | El cerebro |
| **Autonomous Ops** | AutoCompliance, AuditBot | Operaciones autonomas de cumplimiento y control | El cerebro + auditor |

### Priorizacion estrategica

**AgentGuard es el mas importante estrategicamente**: es horizontal (cualquier industria), es urgente (regulacion viene), y tiene el moat mas fuerte (razonamiento formal vs prompts fragiles). Los vertical agents son mas faciles de vender pero mas faciles de replicar.

### Relacion con productos anteriores (motores)

| Motor (docs/) | Agente (este doc) | Evolucion |
|---|---|---|
| ComplianceEngine | AutoCompliance | Motor → agente autonomo |
| LendingBrain | CreditAgent | Motor → agente conversacional |
| UnderwriteIQ | (futuro: UnderwriteAgent) | Motor → agente de suscripcion |
| GovRules | (futuro: GovAgent) | Motor → agente de politica publica |
| ClinicalRules | (futuro: ClinicalAgent) | Motor → agente de soporte clinico |
| AgroDecide | WhatsApp Advisor pattern | Motor → agente via mensajeria |
| — | AgentGuard | NUEVO — governance horizontal |
| — | PolicyForge | NUEVO — diseno de politicas con simulacion causal |
| — | AuditBot | NUEVO — auditoria autonoma |
| — | ClaimsAgent | NUEVO — gestion de siniestros |
