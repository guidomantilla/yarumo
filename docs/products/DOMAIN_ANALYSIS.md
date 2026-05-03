# Analisis de Dominio: Inferential Decision-Making por Vertical

Objetivo: evaluar si "el mejor paquete Go para inferential decision-making que sea profesional, respetable, y completo" cubre las necesidades de dominios de alto valor, e identificar oportunidades de negocio/producto.

Referencia: `modules/compute/PLAN.md` — plan maestro de ejecucion.

---

## 1) IoT (Internet of Things)

### Cobertura: 80%

### Lo que SI cubre el plan

| Necesidad IoT | Que lo resuelve |
|---|---|
| "Temperatura es ALTA y vibracion EXCESIVA → alerta" | Fuzzy engine |
| "Si temperatura > 80°C AND presion > 150psi → shutdown" | Deductive engine |
| "Probabilidad de falla dado estos sintomas?" | Bayesian engine |
| "Si 3 anomalias EN 5 minutos → escalamiento" | Temporal acotado (Fase 2) |
| "La tasa de errores esta cambiando?" | RunningStats Welford |
| "Tiempo hasta fallo de este componente" | Weibull (Fase 1) |
| "Conteo de eventos por unidad de tiempo" | Poisson (Fase 1) |
| "Que pasa si bajo el umbral de alerta?" | Causal (Fase 3) |

### Lo que NO cubre

| Gap | Que es | Impacto |
|---|---|---|
| **Windowed aggregation** | "Promedio de temperatura en los ultimos 5 minutos", "maximo en ventana deslizante de 1 hora" | **Medio-Alto** |
| CEP completo (Complex Event Processing) | Pattern matching sobre streams de eventos con operadores de secuencia | Medio — temporal acotado cubre el 60% |
| Time-series (autocorrelacion, estacionalidad, ARIMA) | Analisis de series temporales | Bajo — scope de data science |
| State machines / automatas | Estados de dispositivo (idle→warming→running→error) | Medio — propuesto como `fsm/` en ROADMAP_MATH.md |

### Gap principal: Windowed Aggregation

RunningStats calcula media/varianza acumulada sobre TODO el stream. IoT necesita ventanas: "los ultimos N puntos" o "los ultimos T minutos". Es la diferencia entre `RunningStats` y `WindowedStats`.

Es inferencia? Discutible. Es infraestructura de datos que alimenta la inferencia. Pero un `WindowedStats` (~80-100 lineas extra sobre RunningStats) haria el paquete mucho mas util para IoT sin salirse del scope.

### Veredicto

80% cubierto. El 20% faltante es un WindowedStats que podria ser nice-to-have.

### Preguntas abiertas

- WindowedStats: ventana por conteo (N puntos) o por tiempo (T minutos) o ambas?
- CEP: el temporal acotado (BEFORE/AFTER/WITHIN) es suficiente para los casos IoT tipicos? O necesita operadores de secuencia (FOLLOWED_BY, NOT_FOLLOWED_BY)?
- ~~State machines: se resuelve fuera (consumidor) o hay un mini-FSM util para IoT dentro del scope?~~ → Propuesto como `math/fsm/` en ROADMAP_MATH.md

---

## 2) Teoria de Operaciones (Operations Research)

### Cobertura: 40%

### Lo que SI cubre el plan

| Necesidad OR | Que lo resuelve |
|---|---|
| "Si inventario < punto_reorden → reabastecer" | Deductive engine |
| "Cual es la probabilidad de stockout?" | Distribuciones (Normal, Poisson) |
| "Actualizar forecast con datos nuevos" | Bayesian engine |
| "Priorizar pedidos con criterios difusos" | Fuzzy engine |
| "Que pasa si cambio la politica de inventario?" | Causal (Fase 3) |
| "El nuevo proceso es significativamente mejor?" | Hypothesis testing (Fase 1) |
| "Tasa de llegada de clientes" | Poisson |
| "Tiempo entre fallas / tiempo de servicio" | Exponential/Weibull |

### Lo que NO cubre

| Gap | Que es | Impacto |
|---|---|---|
| **Optimizacion (LP/IP)** | "Minimizar costo sujeto a restricciones" — el corazon de OR | **Critico para OR puro** |
| Scheduling | Job shop, asignacion de recursos | Alto |
| Queuing theory (M/M/1, M/M/c) | Modelos de colas, capacidad | Medio |
| **MCDM (AHP, TOPSIS, ELECTRE)** | Decision multi-criterio formal | **Medio-Alto** |
| Simulation (DES) | Simulacion de eventos discretos | Medio |
| Markov chains | Procesos estocasticos | Medio-Bajo — propuesto como `markov/` en ROADMAP_MATH.md |

### Gap fundamental

OR es sobre **OPTIMIZACION** (encontrar la MEJOR decision). El plan es sobre **INFERENCIA** (razonar sobre que decision tomar dada evidencia). Son complementarios pero distintos.

Un ingeniero de OR diria: "Tu paquete me ayuda a EVALUAR una decision, pero no a ENCONTRAR la optima."

### Gap recuperable: MCDM

MCDM (Multi-Criteria Decision Making) SI es inferencia sobre decisiones. AHP y TOPSIS son metodos para razonar sobre alternativas dado multiples criterios con pesos. Es lo que un gerente de operaciones usa para decidir entre proveedores, priorizar proyectos, o seleccionar ubicaciones.

No es optimizacion — es razonamiento estructurado sobre alternativas. Encaja en el scope.

### Veredicto

40% cubierto. El paquete cubre el lado de "decision support" de OR (evaluar, razonar, explicar), pero NO el lado de optimizacion. MCDM seria un nice-to-have legitimo dentro del scope.

### Preguntas abiertas

- MCDM: AHP + TOPSIS son suficientes? O necesita ELECTRE/PROMETHEE tambien?
- Queuing theory: un M/M/1 basico (~50 lineas, usa Exponential + Poisson existentes) seria util como ejemplo/utility sin ser un subpaquete?
- El paquete deberia tener una posicion explicita sobre "no somos optimizacion, somos inferencia"?

---

## 3) Agentic AI

### Cobertura: 85%

### Lo que SI cubre el plan

| Necesidad Agentic | Que lo resuelve |
|---|---|
| "El agente puede ejecutar esta accion?" (guardrails) | Deductive engine → ActionGuard |
| "Cual es la confianza en esta evidencia?" | Bayesian engine |
| "Priorizar acciones con incertidumbre" | Fuzzy engine |
| "Que pasa si el agente toma accion X?" | Causal (Fase 3) |
| "Por que el agente tomo esta decision?" | Explain/ across all engines |
| "La decision cumple compliance?" | Validate/ + Audit/ |
| "PARA TODO recurso: recurso.disponible" | Predicados acotados (Fase 2) |
| "SI NO respuesta EN timeout → retry" | Temporal acotado (Fase 2) |
| "Habria sido diferente si el agente hubiera tenido mas datos?" | Contrafactuales (Fase 3) |

### Lo que NO cubre

| Gap | Que es | Impacto |
|---|---|---|
| Planning (STRIPS, HTN, PDDL) | Generacion automatica de planes multi-paso | **Bajo** — el LLM planea, el engine evalua |
| Utility maximization / VOI | "Cuanto vale obtener mas informacion antes de decidir?" | Medio — interesante pero nicho |
| BDI architecture | Belief-Desire-Intention formal | Bajo — el LLM es el BDI implicito |
| Multi-agent coordination | Protocolos de negociacion entre agentes | Bajo — scope de framework, no de inferencia |
| Reinforcement learning | Aprendizaje por recompensa | Cero — scope de ML |
| Ontologias (OWL, RDF) | Razonamiento semantico | Bajo — el LLM lo hace mejor |

### Por que agentic es el dominio mas alineado

La arquitectura moderna de agentes (LangChain, CrewAI, AutoGen) tiene esta division:

```
LLM → PLANEA y RAZONA en lenguaje natural
Inference Engine → EVALUA, RESTRINGE, EXPLICA formalmente
```

El LLM es bueno planificando y generando. Es MALO siendo determinista, auditable, y explicable. Ahi es donde entra el inference engine como **guardrail formal + explicabilidad + auditoria**. El plan cubre exactamente eso.

### Veredicto

85% cubierto. Los gaps son responsabilidad del LLM o del framework agentic, no del engine de inferencia.

### Preguntas abiertas

- Utility/VOI: "cuanto vale obtener mas info antes de decidir?" es inferencial. Util para agentes que deciden si buscar mas datos o actuar ya. Entra en scope?
- El SDK necesita un adapter pattern explicito para frameworks agenticos (tool interface)?
- Multi-agent: si dos agentes consultan el mismo engine, necesita ser thread-safe? (probablemente ya lo es, pero verificar)

---

## 4) Resumen ejecutivo

| Dominio | Cobertura | Gap principal | Entra en scope? |
|---|---|---|---|
| **IoT** | 80% | WindowedStats | Si — extension menor de RunningStats |
| **Operaciones** | 40% | Optimizacion (LP/IP), MCDM | LP/IP: No. MCDM: Si, es inferencia |
| **Agentic AI** | 85% | Planning algorithms | No — es responsabilidad del LLM |

### Adiciones potenciales al plan (nice-to-haves)

| Item | Lineas est. | Dominio que desbloquea | Es inferencia? |
|---|---|---|---|
| WindowedStats | ~100 | IoT | Discutible — alimenta inferencia |
| MCDM basico (AHP + TOPSIS) | ~300-400 | OR, tambien util en agentic | Si — razonamiento multi-criterio |

Ninguno cambia la arquitectura. Ninguno bloquea el claim actual. Ambos amplian el alcance a dominios de alto valor sin salirse de "inferential decision-making".

---

## 5) Posicionamiento: que ES y que NO ES el paquete

Para evitar confusion, el paquete deberia tener una posicion explicita:

**ES:**
- Motor de inferencia formal (reglas, probabilidad, fuzzy, causalidad)
- Soporte de decision (evaluar, explicar, auditar)
- Razonamiento multi-criterio (si se agrega MCDM)
- Substrato formal para agentes AI

**NO ES:**
- Optimizador (LP/IP/scheduling) — usar OR-Tools, GLPK, o gonum
- Framework de ML/AI — usar goml, gorgonia, o APIs de LLM
- CEP/stream processor — usar Apache Flink, Kafka Streams
- Theorem prover — usar Z3, Lean, Coq
- Framework agentico — usar LangChain, CrewAI, Aluna

---

## 6) Analisis extendido: mas verticales

### Criterio de seleccion

El sweet spot del paquete es donde convergen:
1. **Multiples tipos de razonamiento** — no basta con reglas simples
2. **Explicabilidad obligatoria** — regulador, auditor, o usuario exige "por que"
3. **Decisiones de alto impacto** — dinero, salud, libertad, reputacion en juego
4. **Compliance formal** — deadlines legales, trazabilidad, auditoria

Si un dominio solo necesita 1 de los 4, probablemente le basta con un motor de reglas simple. Si necesita 3-4, el paquete es exactamente lo que necesita.

---

### 6a) FinTech / Lending

**Cobertura estimada: 90%** — el dominio mas natural

| Necesidad | Paradigma | Ejemplo concreto |
|---|---|---|
| Reglas de politica crediticia | Deductive | "Si score < 600 AND antiguedad < 1 ano → rechazar" |
| Probabilidad de default | Bayesian | P(default \| ingresos, deuda, historial) |
| Score difuso de riesgo | Fuzzy | "riesgo es ALTO cuando ingreso es BAJO y deuda es ALTA" |
| "Que pasa si bajo el umbral de score?" | Causal | Impacto en tasa de aprobacion vs tasa de mora |
| Seleccion de producto financiero | MCDM | "Cual credito ofrecer dados tasa, plazo, riesgo, rentabilidad" |
| "Por que se rechazo?" | Explain/ | Obligatorio por ley (habeas data, fair lending) |
| Deadlines regulatorios | Temporal | PQR: responder en 15 dias habiles (Ley 1755/2015) |
| Auditoria de decisiones | Audit/ | Superintendencia Financiera, SARLAFT |

**Gaps**: scoring models pre-entrenados (ML, fuera de scope), integracion con buros de credito (infra).

**Producto**: **LendingBrain** — motor de decisiones para originacion de credito. El banco define reglas, el motor decide, explica, y audita. Diferenciador vs competencia (Mambu, nCino): explicabilidad formal + causal "what-if" para calibracion de politicas.

**Mercado LATAM**: Bancos, fintech, cooperativas. Regulacion SFC obliga explicabilidad. ~200+ fintech en Colombia, ~800+ en LATAM.

---

### 6b) InsurTech / Underwriting

**Cobertura estimada: 85%**

| Necesidad | Paradigma | Ejemplo concreto |
|---|---|---|
| Reglas de suscripcion | Deductive | "Si edad > 65 AND condicion preexistente → exclusion" |
| Probabilidad de siniestro | Bayesian | P(claim \| perfil, historial, zona) |
| Evaluacion difusa de riesgo | Fuzzy | "riesgo vehicular es MEDIO cuando zona es MODERADA y historial es BUENO" |
| "Que pasa si incluimos esta cobertura?" | Causal | Impacto en siniestralidad esperada |
| Seleccion de reaseguro | MCDM | "Cual reasegurador dados precio, solvencia, respuesta, cobertura" |
| Tiempo hasta siniestro | Weibull | Modelado de confiabilidad, reservas |
| Auditoria actuarial | Audit/ | Superfinanciera, FASECOLDA |

**Gaps**: modelos actuariales complejos (GLM, fuera de scope), integracion con sistemas legacy (infra).

**Producto**: **UnderwriteIQ** — motor de suscripcion automatizada. La aseguradora define politicas, el motor evalua riesgo con multiples paradigmas y explica cada decision. Diferenciador: no es black-box ML, es razonamiento formal auditable.

**Mercado**: aseguradoras, insurtech, corredores. Regulacion exige explicabilidad en rechazos.

---

### 6c) HealthTech / Clinical Decision Support

**Cobertura estimada: 80%**

| Necesidad | Paradigma | Ejemplo concreto |
|---|---|---|
| Protocolos clinicos | Deductive | "Si fiebre > 38.5 AND tos > 7 dias AND rx anormal → sospecha neumonia" |
| Diagnostico diferencial | Bayesian | P(enfermedad \| sintomas, examenes, demograficos) |
| Triage difuso | Fuzzy | "urgencia es ALTA cuando dolor es SEVERO y signos vitales INESTABLES" |
| "Que pasa si cambiamos el protocolo?" | Causal | Impacto en outcomes, falsos negativos |
| Seleccion de tratamiento | MCDM | "Cual tratamiento dados eficacia, efectos secundarios, costo, adherencia" |
| Alertas temporales | Temporal | "Si no mejora EN 48 horas → escalar a especialista" |
| Auditoria clinica | Audit/ | Habilitacion, acreditacion, Supersalud |

**Gaps**: interoperabilidad HL7/FHIR (infra), base de conocimiento medico (datos), aprobacion regulatoria medica (proceso largo).

**Producto**: **ClinicalRules** — motor de soporte a decision clinica embebible. Hospitales/EPS definen protocolos formalmente, el motor evalua y explica. NO reemplaza al medico — lo asiste con razonamiento formal y trazabilidad.

**Mercado LATAM**: hospitales, EPS, IPS. La Ley Estatutaria de Salud exige calidad y trazabilidad. RIPS (Registro Individual de Prestacion de Servicios) exige auditabilidad.

**Precaucion**: regulacion de dispositivos medicos (INVIMA en Colombia, FDA en US). Un motor de "soporte" es diferente de un motor de "diagnostico". Posicionarse como soporte, no como dispositivo medico.

---

### 6d) LegalTech / RegTech

**Cobertura estimada: 85%**

| Necesidad | Paradigma | Ejemplo concreto |
|---|---|---|
| Reglas regulatorias | Deductive | "Si monto > 10M AND cliente PEP → reporte UIAF" |
| Probabilidad de riesgo legal | Bayesian | P(sancion \| hallazgo, antecedentes, jurisdiccion) |
| Severidad difusa de incumplimiento | Fuzzy | "riesgo es CRITICO cuando frecuencia ALTA y impacto ALTO" |
| "Que pasa si cambia la regulacion?" | Causal | Impacto de nueva circular SFC en portafolio |
| Priorizacion de hallazgos | MCDM | "Cual hallazgo atender primero dados impacto, urgencia, costo, reputacion" |
| Plazos legales | Temporal | Tutela: 10 dias (D.2591/91), PQR: 15 dias habiles |
| Evidencia de cumplimiento | Audit/ | Superintendencias, Contraloria, Procuraduria |
| Cuantificadores sobre documentos | Predicados | "PARA TODO contrato EN portafolio: contrato.clausula_penal EXISTS" |

**Gaps**: NLP para extraccion de reglas de texto legal (LLM, fuera de scope), jurisprudencia (datos).

**Producto**: **ComplianceEngine** — motor de cumplimiento regulatorio. Define regulaciones como reglas formales, monitorea plazos, prioriza riesgos, genera evidencia de cumplimiento. Diferenciador: temporal logic para plazos legales LATAM + auditoria completa.

**Mercado LATAM**: bancos (SARLAFT/SAGRILAFT), empresas de servicios publicos (SSPD), salud (Supersalud), cualquier entidad vigilada. Mercado de compliance en LATAM crece ~15% anual.

---

### 6e) AgriTech / Smart Agriculture

**Cobertura estimada: 75%**

| Necesidad | Paradigma | Ejemplo concreto |
|---|---|---|
| Reglas de manejo de cultivo | Deductive | "Si humedad suelo < 30% AND pronostico sin lluvia → regar" |
| Probabilidad de plaga | Bayesian | P(roya \| humedad, temperatura, historial, vecinos) |
| Evaluacion difusa de condiciones | Fuzzy | "salud del cultivo es REGULAR cuando color AMARILLENTO y crecimiento LENTO" |
| "Que pasa si cambio la densidad de siembra?" | Causal | Impacto en rendimiento por hectarea |
| Seleccion de cultivo/variedad | MCDM | "Cual variedad dados rendimiento, resistencia, precio, ciclo, agua" |
| Alertas IoT | Temporal + WindowedStats | "Si temperatura promedio > 35°C EN 3 dias → alerta estres termico" |

**Gaps**: modelos agronomicos especializados (datos), imagenes satelitales/drones (CV, fuera de scope), datos meteorologicos (integracion externa).

**Producto**: **AgroDecide** — motor de decision para agricultura de precision. El agronomo define reglas de manejo, sensores IoT alimentan el motor, el motor decide y explica. Diferenciador: fuzzy (condiciones de campo no son binarias) + temporal (patrones climaticos) + explicabilidad (el agricultor entiende POR QUE).

**Mercado LATAM**: Colombia es 5to productor mundial de cafe, 1ro en flores, gran productor de palma, banano, cana. Cafeteros (~500K familias) con problemas de roya, broca. Floricultura de exportacion con control ambiental critico.

---

### 6f) GovTech / Government Decision-Making

**Cobertura estimada: 80%**

| Necesidad | Paradigma | Ejemplo concreto |
|---|---|---|
| Reglas de elegibilidad | Deductive | "Si SISBEN < X AND edad > 60 AND no pensionado → subsidio" |
| Focalizacion probabilistica | Bayesian | P(vulnerable \| indicadores socioeconomicos) |
| Priorizacion difusa | Fuzzy | "urgencia es ALTA cuando pobreza EXTREMA y acceso BAJO" |
| "Que pasa si cambiamos el umbral de SISBEN?" | Causal | Impacto en cobertura vs presupuesto |
| Priorizacion de proyectos | MCDM | "Cual proyecto dados impacto social, costo, viabilidad, poblacion beneficiada" |
| Plazos legales | Temporal | Derechos de peticion, tutelas, PQRS |
| Transparencia | Explain/ + Audit/ | Ley de Transparencia (1712/2014), rendicion de cuentas |

**Gaps**: integracion con sistemas gubernamentales legacy (SIIF, SECOP), volumen masivo de beneficiarios (escala).

**Producto**: **GovRules** — motor de decision para politica publica. Define reglas de focalizacion, elegibilidad, y priorizacion formalmente. Cada decision explicada y auditable. Diferenciador: transparencia formal (no es "el algoritmo decidio" sino "estas son las reglas y asi se aplicaron").

**Mercado LATAM**: DNP, ministerios, alcaldias, gobernaciones. Colombia tiene ~1,100 municipios. Programa Colombia Potencia de la Vida impulsa transformacion digital gubernamental.

---

## 7) Ideas de producto concretas — priorizadas

### Criterio de priorizacion

| Factor | Peso |
|---|---|
| Encaje con el paquete (% cobertura) | Alto |
| Tamano de mercado LATAM | Alto |
| Obligacion regulatoria de explicabilidad | Alto |
| Barrera de entrada tecnica | Medio |
| Complejidad de datos externos | Medio (penaliza) |

### Ranking

| # | Producto | Vertical | Encaje | Mercado | Regulacion | Complejidad datos | Score |
|---|---|---|---|---|---|---|---|
| 1 | **ComplianceEngine** | RegTech | 85% | Grande | Obligatoria | Baja (reglas, no ML) | ★★★★★ |
| 2 | **LendingBrain** | FinTech | 90% | Grande | Obligatoria | Media (buros) | ★★★★★ |
| 3 | **UnderwriteIQ** | InsurTech | 85% | Grande | Obligatoria | Media (actuarial) | ★★★★☆ |
| 4 | **GovRules** | GovTech | 80% | Grande | Obligatoria | Baja (datos publicos) | ★★★★☆ |
| 5 | **ClinicalRules** | HealthTech | 80% | Grande | Obligatoria | Alta (HL7/FHIR) | ★★★☆☆ |
| 6 | **AgroDecide** | AgriTech | 75% | Medio | Baja | Alta (IoT + meteo) | ★★★☆☆ |

### La jugada estrategica

Los productos 1-4 comparten una verdad: **en LATAM, la regulacion OBLIGA explicabilidad**. No es un nice-to-have — es la ley. Superintendencias (SFC, Supersalud, SSPD), Contraloria, Procuraduria, jueces de tutela — todos exigen "por que decidiste esto". Y la respuesta no puede ser "el modelo de ML dijo 0.73".

El paquete no compite con ML. **Complementa ML**: el modelo predice, el motor de inferencia decide, explica, y audita. Son capas diferentes.

### Modelo de negocio posible

**Opcion A: Open-core**
- yarumo (open source) = el motor de inferencia
- Producto comercial = vertical-specific platform (UI + integraciones + soporte + SLA)
- Comparable: Drools (open) → Red Hat Decision Manager (comercial)

**Opcion B: Embedded engine**
- yarumo como dependencia Go en productos de terceros
- Monetizacion via soporte enterprise, certificacion, training
- Comparable: SQLite (embedded, public domain) → empresas pagan por extensiones y soporte

**Opcion C: DaaS (Decision-as-a-Service)**
- yarumo como backend de un servicio cloud multi-tenant
- API REST/gRPC para evaluar decisiones
- Ya documentado en `docs/DAAS_ARCH.md`
- Comparable: AWS Fraud Detector, Google Recommendations AI

**Opcion D: AI Guardrails Platform**
- yarumo como substrato formal para agentes AI
- El producto es "governance para AI agents": define que pueden/no pueden hacer, audita todo
- Mercado emergente: AI governance es el compliance del 2026-2030
- Comparable: Guardrails AI, pero con razonamiento formal en vez de solo LLM-based checks

---

## 8) El mapa completo

```
                        yarumo (inference engine)
                                |
                    +-----------+-----------+
                    |                       |
              Producto vertical        Plataforma horizontal
                    |                       |
        +-----------+----------+    +-------+-------+
        |           |          |    |               |
   LendingBrain  GovRules  ...   DaaS         AI Guardrails
   (FinTech)    (GovTech)       (multi-tenant)  (Agentic)
        |           |          |    |               |
        +-----------+----------+    +-------+-------+
                    |                       |
                    +--- Aluna (agentic platform) ---+
                         (orquesta todo)
```

Aluna no ES ninguno de estos productos — Aluna los ORQUESTA. El motor de inferencia es la pieza de razonamiento formal que Aluna (o cualquier plataforma agentica) usa para decidir, explicar, y auditar.

La pregunta no es "cual producto hacer primero". Es: **terminar el motor (PLAN.md Fases 0-3), luego elegir la vertical con mas traccion**. El motor es el asset reutilizable. Los productos son las instancias.
