# ComplianceEngine — Motor de Cumplimiento Regulatorio

Vertical: RegTech. Encaje con yarumo: 85%. Score: ★★★★★.

Referencia: `docs/DOMAIN_ANALYSIS.md` seccion 6d y seccion 7.

---

## 1) El problema

El oficial de cumplimiento de un banco/EPS/empresa de servicios publicos hoy trabaja con:

- **Excel** para rastrear plazos regulatorios
- **Email** para evidencia de cumplimiento
- **Memoria humana** para saber que regulacion aplica a que
- **Panico** cuando la Superintendencia pide "demuestreme que cumple"

Cuando la SFC pregunta "por que le negaron el credito a este cliente?", la respuesta no puede ser "el modelo dijo 0.73". Necesitan trazabilidad formal.

---

## 2) Que hace

ComplianceEngine formaliza regulaciones como reglas ejecutables, monitorea plazos, evalua riesgo de incumplimiento, prioriza hallazgos, y genera evidencia auditable automaticamente.

---

## 3) Como usa cada paradigma

| Paradigma | Pregunta que responde | Ejemplo concreto |
|---|---|---|
| **Deductive** | "Esta regulacion aplica a esta entidad?" | "Si entidad es vigilada SFC AND activos > 10.000 SMLMV → aplica SARLAFT" |
| **Bayesian** | "Cual es la probabilidad de sancion?" | P(sancion \| hallazgo recurrente, antecedentes, tipo de entidad) |
| **Fuzzy** | "Que tan grave es este incumplimiento?" | "severidad es CRITICA cuando frecuencia ALTA y impacto ALTO y visibilidad PUBLICA" |
| **Causal** | "Si corregimos este proceso, baja el riesgo?" | Impacto de implementar control X en probabilidad de sancion Y |
| **MCDM** | "Cual hallazgo atender primero?" | Ranking dados: impacto economico, plazo, probabilidad de deteccion, costo de remediacion |
| **Temporal** | "Estamos dentro del plazo legal?" | PQR: 15 dias habiles (Ley 1755). Tutela: 10 dias (D.2591). Habeas data: 15 dias (Ley 1581) |
| **Predicados** | "Todos los contratos tienen clausula X?" | `PARA TODO contrato EN portafolio: contrato.clausula_penal EXISTS` |
| **Explain/Audit** | "Por que se tomo esta decision?" | Trace completo: que reglas se evaluaron, con que datos, que resultado, cuando |

---

## 4) Flujo tipico

```
1. Oficial de cumplimiento define regulaciones como reglas formales
   (o LLM las extrae del texto legal → NLP companion module)

2. Sistema recibe eventos:
   - PQR ingresada (inicia conteo de plazo)
   - Reporte SARLAFT vencido
   - Nueva circular SFC publicada
   - Hallazgo de auditoria interna

3. Motor evalua:
   Deductive  → aplica? que obligaciones se activan?
   Temporal   → hay plazo corriendo? cuanto queda?
   Bayesian   → cual es el riesgo de sancion?
   Fuzzy      → que tan severo es?
   MCDM       → prioridad vs otros hallazgos?
   Causal     → que controles mitigarian esto?

4. Output:
   - Dashboard de estado de cumplimiento
   - Alertas de plazos proximos a vencer
   - Ranking priorizado de gaps
   - Reporte auditable para Superintendencia
   - Evidencia formal de cada decision tomada
```

---

## 5) Regulaciones LATAM que cubre naturalmente

| Regulacion | Pais | Que exige | Que paradigma lo resuelve |
|---|---|---|---|
| SARLAFT/SAGRILAFT | CO | AML: reporte de operaciones sospechosas | Deductive (reglas) + Bayesian (riesgo) + Temporal (plazos UIAF) |
| Ley 1755/2015 | CO | PQR: respuesta en 15 dias habiles | Temporal + Deductive |
| Decreto 2591/1991 | CO | Tutela: fallo en 10 dias | Temporal + Deductive + Explain (juez exige motivacion) |
| Ley 1581/2012 | CO | Habeas data: respuesta en 15 dias | Temporal + Deductive |
| Circular Basica Juridica SFC | CO | Gobierno corporativo, gestion de riesgos | Todo — es un marco completo |
| Ley 1328/2009 | CO | Defensor del consumidor financiero | Temporal + Explain + Audit |
| Resolucion 4100/2012 Supersalud | CO | Calidad en salud, tiempos de atencion | Temporal + Fuzzy (calidad) |
| Ley Fintech (2018) | MX | Regulacion de instituciones de tecnologia financiera | Deductive + Bayesian + Audit |
| LGPD | BR | Proteccion de datos (similar GDPR) | Deductive + Temporal + Explain |
| Ley 19.628 / 21.096 | CL | Proteccion de datos Chile | Deductive + Temporal + Explain |

---

## 6) Mercado

### Quien paga

| Segmento | Tamano CO | Dolor | Presupuesto |
|---|---|---|---|
| Bancos | ~30 entidades | Multas SFC (hasta 1% de activos) | Alto |
| Aseguradoras | ~45 entidades | Multas SFC + FASECOLDA | Alto |
| Cooperativas financieras | ~180 entidades | Mismo regimen SFC, menos recursos | Medio |
| EPS/IPS | ~50 EPS, ~10K IPS | Supersalud, sanciones + intervencion | Medio |
| Servicios publicos | ~300 empresas | SSPD, sanciones | Medio |
| Fintech | ~200+ entidades | SFC sandbox, en proceso de regulacion | Medio (creciendo) |

### Multas reales recientes (Colombia)

- SFC a Bancolombia: $1.200M COP (2023) por SARLAFT
- Supersalud a EPS: intervenciones y liquidaciones
- SIC por habeas data: hasta 2.000 SMLMV (~$2.600M COP)

El costo de NO cumplir es altisimo. ComplianceEngine se vende como seguro, no como software.

---

## 7) Competencia

| Competidor | Precio | Fortaleza | Debilidad vs ComplianceEngine |
|---|---|---|---|
| IBM OpenPages | $$$$ | Enterprise completo | Caro, no tiene inferencia formal, no entiende regulacion LATAM |
| SAP GRC | $$$$ | Integracion SAP | Mismo — caro, generico |
| MetricStream | $$$ | Cloud-native | No tiene razonamiento causal ni fuzzy |
| Pirani (CO) | $$ | Local, conoce regulacion CO | Solo gestion de riesgos, no motor de inferencia |
| Excel + SharePoint | $ | "Ya lo tenemos" | No escala, no auditable, no explainable |

**Diferenciador**: ninguno tiene razonamiento formal multi-paradigma con explicabilidad integrada. Todos son o gestion documental glorificada o ML black-box. ComplianceEngine razona, explica, y audita formalmente.

---

## 8) Stack tecnico

```
┌─────────────────────────────────────────┐
│          Frontend (dashboard)            │
├─────────────────────────────────────────┤
│          API REST/gRPC (DaaS)           │
├─────────────────────────────────────────┤
│     SDK decision-engine/core/            │
│  ┌─────────┬──────────┬───────────────┐ │
│  │decisions│ validate │    audit      │ │
│  │ explain │repository│              │ │
│  └─────────┴──────────┴───────────────┘ │
├─────────────────────────────────────────┤
│        yarumo compute/                 │
│  ┌────────┬─────────┬───────┬─────────┐ │
│  │deduct. │bayesian │ fuzzy │ causal  │ │
│  │        │         │       │  mcdm   │ │
│  └────────┴─────────┴───────┴─────────┘ │
├─────────────────────────────────────────┤
│        yarumo maths/                     │
│  ┌──────┬───────┬──────┬──────────────┐ │
│  │logic │ fuzzy │ sets │    stats     │ │
│  └──────┴───────┴──────┴──────────────┘ │
├─────────────────────────────────────────┤
│   NLP companion (LLM: texto→reglas)     │
│   Storage (postgres/redis)              │
│   Telemetry (otel)                      │
└─────────────────────────────────────────┘
```

Todo lo de abajo ya existe o esta en el plan. Lo que faltaria construir es la capa de arriba: frontend + API + integraciones especificas por regulacion.

---

## 9) Go-to-market

**Fase 1**: Un banco o fintech como design partner. Implementar SARLAFT + PQR. Demostrar valor con una auditoria de SFC real.

**Fase 2**: Expandir a 3-5 entidades financieras. Agregar mas regulaciones (habeas data, consumidor financiero).

**Fase 3**: Saltar a otro sector (salud o servicios publicos) con el mismo motor, diferentes reglas.

**Fase 4**: Regionalizar — Mexico (CNBV), Brasil (BACEN), Chile (CMF). Misma arquitectura, diferentes regulaciones.

---

## 10) Por que 85% y no 100%

El 15% faltante:

| Gap | Que es | Como se resuelve |
|---|---|---|
| NLP: texto legal → reglas | Extraer reglas de circulares/leyes automaticamente | LLM companion module (ya en plan SDK) |
| Gestion documental | Almacenar evidencia adjunta | Storage companion module (ya en plan SDK) |
| Calendario habil | Dias habiles varian por jurisdiccion/sector | Utility pequeno o integracion externa |
| Notificaciones | Alertas por email/Slack/Teams | Infra, no inferencia |

Ninguno es un gap del motor de inferencia. Son capas de producto que se construyen encima.
