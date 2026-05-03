# UnderwriteIQ — Motor de Suscripcion Automatizada para Seguros

Vertical: InsurTech / Underwriting. Encaje con yarumo: 85%. Score: ★★★★☆.

Referencia: `docs/DOMAIN_ANALYSIS.md` seccion 6b (InsurTech) y seccion 7.

---

## 1) El problema

Una aseguradora que suscribe polizas hoy tiene:

- **Tablas actuariales estaticas** que se recalculan una vez al ano — el mundo cambia mas rapido
- **Underwriting manual** donde el suscriptor revisa expedientes con criterio personal (inconsistente entre personas, turnos, y sucursales)
- **Modelos actuariales en Excel/R/SAS** que nadie mas entiende — el actuario se va y el modelo se vuelve caja negra
- **Apetito de riesgo declarativo** — el comite dice "no queremos riesgo alto en auto Bogota" pero no hay forma de verificar si las decisiones reales lo cumplen
- **Cero contrafactuales** — "habriamos aceptado este riesgo con la politica anterior?" → nadie sabe
- **Reaseguro por intuicion** — seleccion de reasegurador basada en relacion comercial, no en analisis multi-criterio formal

Cuando la SFC o FASECOLDA pregunta "como calcularon la prima?" o "por que rechazaron esta poliza?", la respuesta es arqueologia documental.

---

## 2) Que hace

UnderwriteIQ es el motor que evalua riesgo asegurable, decide aceptar/rechazar/condicionar, calcula prima ajustada, selecciona condiciones optimas, y genera trazabilidad completa. No reemplaza al actuario — formaliza su conocimiento en reglas ejecutables y auditables.

---

## 3) Como usa cada paradigma

| Paradigma | Pregunta que responde | Ejemplo concreto |
|---|---|---|
| **Deductive** | "Este riesgo es asegurable?" | "Si edad > 75 AND actividad = mineria subterranea → declinar. Si vehiculo > 15 anos AND uso comercial → exclusion" |
| **Bayesian** | "Cual es la probabilidad real de siniestro?" | P(siniestro \| zona, historial, tipo vehiculo, perfil conductor) — actualiza priors con siniestralidad propia |
| **Fuzzy** | "Cual es el nivel de riesgo integral?" | "riesgo es ALTO cuando zona PELIGROSA y vehiculo ANTIGUO y conductor JOVEN" — el riesgo no es binario |
| **Causal** | "Que pasa si cambiamos la politica?" | "Si incluimos cobertura de terremoto en la zona: +8% siniestralidad esperada, +$X en reservas, necesidad de reaseguro facultativo" |
| **MCDM** | "Cual reasegurador elegir?" | Ranking dados: precio, solvencia, velocidad de respuesta, cobertura, relacion historica, rating AM Best |
| **Temporal** | "Plazos regulatorios y contractuales?" | "Si no respuesta a reclamacion EN 30 dias → silencio administrativo positivo". Vencimiento de poliza, periodo de gracia |
| **Predicados** | "Expediente de suscripcion completo?" | `PARA TODO requisito EN producto.requisitos: expediente.tiene(requisito) = true` |
| **Explain/Audit** | "Por que esta prima?" | "Prima base: $1.2M. Ajuste zona +15% (bayesian: zona alta siniestralidad). Ajuste historial -8% (sin reclamos 3 anos). Ajuste edad +5% (fuzzy: conductor joven). Prima final: $1.34M. Reglas aplicadas: R1, R4, R7" |
| **Stats** | "Cuando fallara este equipo?" | Weibull para tiempo hasta fallo en seguros de ingenieria. Poisson para frecuencia de siniestros. Exponential para tiempo entre reclamaciones |

---

## 4) Lineas de negocio que cubre

| Linea | Decisiones clave | Paradigmas dominantes |
|---|---|---|
| **Auto** | Aceptar/rechazar, prima, deducible, coberturas | Deductive + Bayesian + Fuzzy |
| **Vida individual** | Aceptar/rechazar/condicionar, prima por edad-salud | Deductive + Bayesian + Temporal (mortalidad) |
| **Vida grupo** | Tarifar colectivo, seleccion de riesgo grupal | Bayesian + Stats (distribuciones) + Predicados (requisitos grupo) |
| **Hogar/Propiedad** | Evaluar riesgo zona, construccion, coberturas | Fuzzy (riesgo zona) + Deductive + Causal (catastrofico) |
| **SOAT** | Regulado, pero ajuste de reservas y fraude | Deductive (tarifa regulada) + Bayesian (fraude) |
| **Cumplimiento/Fianzas** | Evaluar solvencia del afianzado | MCDM + Bayesian + Predicados (documentos financieros) |
| **Ingenieria/Todo riesgo** | Evaluar proyecto, tiempo de exposicion | Weibull + Fuzzy + Causal (impacto de medidas de mitigacion) |
| **Reaseguro** | Seleccion de reasegurador, cesion optima | MCDM (multi-criterio) + Stats (distribuciones de perdida) |

---

## 5) Flujo de suscripcion

```
1. Solicitud de poliza (corredor/canal digital/sucursal)

2. Datos entran al motor:
   - Informacion del asegurado (persona/empresa)
   - Objeto asegurado (vehiculo, propiedad, vida, proyecto)
   - Historial de siniestralidad (interna + FASECOLDA)
   - Inspeccion (si aplica)
   - Condiciones solicitadas (coberturas, limites, deducibles)

3. Motor evalua en cascada:

   Predicados  → Expediente completo? Requisitos del producto?
       ↓ si completo
   Deductive   → Riesgo asegurable? (knock-out: exclusiones absolutas)
       ↓ si asegurable
   Bayesian    → P(siniestro) por tipo, actualizada con data propia
       ↓
   Fuzzy       → Riesgo integral (perfil + zona + objeto + historial)
       ↓
   Stats       → Distribucion de perdida esperada (severidad + frecuencia)
       ↓
   Deductive   → Prima = base x ajustes (reglas de tarificacion)
       ↓
   MCDM        → Necesita reaseguro? Cual reasegurador?
       ↓
   Causal      → Sensitivity: que cambiando que condicion cambia la decision?

4. Output:
   - ACEPTAR / DECLINAR / CONDICIONAR / REFERIR A SUSCRIPTOR SENIOR
   - Prima calculada con desglose de ajustes
   - Coberturas recomendadas y exclusiones aplicadas
   - Condiciones especiales (deducible mayor, clausulas, sublimites)
   - Necesidad de reaseguro (automatico/facultativo)
   - Explicacion completa por paradigma
   - Audit trail para SFC / FASECOLDA / auditoria interna
```

---

## 6) El diferenciador: tarificacion explicable

Lo que hace unico a UnderwriteIQ no es calcular una prima — cualquier Excel hace eso. Es **explicar por que esa prima y no otra**:

```
"Poliza auto — Toyota Corolla 2022, conductor M/28 anos, Bogota norte:

 1. Asegurabilidad: ACEPTAR (sin exclusiones absolutas)
 2. P(siniestro/ano) = 0.12 (bayesian):
    - Prior sector: 0.15 (autos sedan Bogota)
    - Ajuste historial: -0.04 (sin siniestros 2 anos)
    - Ajuste edad: +0.01 (conductor < 30)
 3. Riesgo integral: MEDIO (0.45 fuzzy)
    - Zona: MODERADA (0.5) — norte Bogota, robo medio
    - Vehiculo: BAJO (0.2) — nuevo, buen safety rating
    - Conductor: MEDIO (0.55) — joven pero sin historial negativo
 4. Perdida esperada: $2.1M (Poisson freq=0.12, Lognormal sev mu=$4.2M)
 5. Prima pura: $504K. Gastos: +35%. Utilidad: +12%. Prima comercial: $764K
 6. Ajustes aplicados: descuento sin siniestros -8%, recargo conductor joven +5%
 7. Prima final: $741K

 Causal: si el conductor tuviera 35+ anos, prima bajaria a $698K (-5.8%).
 Si trasladara el vehiculo a zona de menor riesgo, bajaria a $672K (-9.3%).
```

Eso lo presenta el corredor al cliente. Eso lo muestra el suscriptor al comite. Eso lo entrega la aseguradora a la SFC.

---

## 7) Mercado

### Quien paga

| Segmento | Tamano CO | Volumen de polizas | Dolor |
|---|---|---|---|
| Aseguradoras grandes | ~10 (Sura, Bolivar, Liberty, Allianz, Mapfre...) | Miles/dia | Consistencia, escala, regulacion |
| Aseguradoras medianas | ~15 | Cientos/dia | Competir con grandes sin equipo actuarial grande |
| Aseguradoras pequenas/nicho | ~20 | Decenas-cientos/dia | Digitalizacion, eficiencia |
| Cooperativas de seguros | ~5 | Cientos/dia | Mismo regimen SFC, menos recursos |
| Corredores de reaseguro | ~15 | Decenas/dia | Seleccion de reasegurador formal |
| Insurtech | Emergente (~10+) | Variable | Escalar suscripcion digital sin suscriptor humano |

### Numeros del mercado

- Primas emitidas Colombia: ~$40B COP (2024)
- LATAM mercado seguros: ~$180B USD
- Cada punto de siniestralidad reducida = millones en resultados tecnicos
- Costo de suscriptor senior: ~$8-15M COP/mes — UnderwriteIQ escala sin headcount

---

## 8) Competencia

| Competidor | Que hace | Debilidad vs UnderwriteIQ |
|---|---|---|
| Guidewire (Underwriting) | Suite completa P&C | $$$$$, monolitico, implementacion 12-18 meses, no causal |
| Duck Creek | Policy + rating engine | $$$$, US-centric, no multi-paradigma |
| Majesco | Cloud P&C | $$$, no tiene fuzzy/causal/MCDM |
| Sapiens | Core insurance + BI | $$$, mas core que decision |
| Willis Towers Watson (Radar) | Pricing actuarial | Actuarial puro, no motor de reglas, no explicabilidad operativa |
| Excel + VBA del actuario | "Asi lo hemos hecho siempre" | Un humano, un bus factor, cero auditoria |

**Diferenciador**: los competidores son o plataformas monoliticas caras (Guidewire, Duck Creek) o herramientas actuariales puras (WTW Radar). Ninguno combina suscripcion + tarificacion + explicabilidad + causal "what-if" en un motor ligero embebible. UnderwriteIQ es el motor, no la plataforma — se integra con el core existente.

---

## 9) Regulacion que lo hace obligatorio

| Norma | Que exige | Como UnderwriteIQ cumple |
|---|---|---|
| Decreto 2555/2010 (regimen seguros) | Reservas tecnicas adecuadas, tarificacion fundamentada | Stats (distribuciones) + Bayesian (frecuencia/severidad) + Audit |
| Circular Basica Juridica SFC (seguros) | Gestion de riesgo de suscripcion | Fuzzy + Bayesian + Deductive, todo trazable |
| Ley 1328/2009 | Proteccion al consumidor financiero | Explain/ (por que esta prima, por que este rechazo) |
| NIIF 17 (IFRS 17) | Medicion de contratos de seguros, best estimate | Stats (distribuciones de perdida) + Bayesian (actualizacion de supuestos) |
| Circular 029/2014 SFC | SARO (riesgo operacional) | Audit/ + Temporal (monitoreo de controles) |
| FASECOLDA estandares | Datos sectoriales, benchmarking | Bayesian (priors sectoriales FASECOLDA como base) |
| Resolucion 3A de URF | Reservas de siniestros | Stats (triangulos de desarrollo → distribuciones) |

**NIIF 17 es el game-changer**: desde 2023, las aseguradoras deben medir contratos con "best estimate" de flujos futuros. Eso requiere distribuciones de probabilidad, no tablas fijas. Stats/ con Poisson (frecuencia) + Lognormal (severidad) + RunningStats (monitoreo) es exactamente lo que NIIF 17 pide.

---

## 10) Stack tecnico

```
┌──────────────────────────────────────────────┐
│  Frontend (portal suscriptor + portal corredor)│
├──────────────────────────────────────────────┤
│  API REST/gRPC                                │
│  ┌───────────────────────────────────────┐    │
│  │ Orchestrator (cascada de suscripcion) │    │
│  │ PreCheck → KnockOut → Risk → Rate →  │    │
│  │ Reinsurance → Sensitivity             │    │
│  └───────────────────────────────────────┘    │
├──────────────────────────────────────────────┤
│  SDK decision-engine/core/                    │
│  (decisions + explain + validate + audit)     │
├──────────────────────────────────────────────┤
│  yarumo compute/ + maths/                   │
│  (deductive, bayesian, fuzzy, causal, mcdm)  │
├──────────────────────────────────────────────┤
│  Integraciones externas                       │
│  ┌────────────┬───────────┬───────────────┐  │
│  │ FASECOLDA  │ Core      │ Inspecciones  │  │
│  │ (datos     │ seguros   │ (fotos, docs, │  │
│  │ sectoriales│ (emision) │  peritaje)    │  │
│  │ RUNT       │           │               │  │
│  └────────────┴───────────┴───────────────┘  │
├──────────────────────────────────────────────┤
│  Storage + Telemetry + NLP companion          │
└──────────────────────────────────────────────┘
```

---

## 11) Go-to-market

**Fase 1**: Una aseguradora mediana como design partner. Linea auto (mayor volumen, datos abundantes). Implementar knock-out rules + bayesian risk + tarificacion explicable. Demostrar consistencia vs suscriptores humanos y generacion automatica de justificacion de prima.

**Fase 2**: Agregar fuzzy risk scoring + MCDM para seleccion de reaseguro. Demostrar que el motor iguala o mejora la seleccion del suscriptor senior en 80%+ de los casos.

**Fase 3**: Expandir a linea vida y hogar. Agregar causal "what-if" para el comite tecnico: "si cambiamos el deducible minimo de $500K a $1M, impacto en siniestralidad y retencion".

**Fase 4**: NIIF 17 compliance module — distribuciones de perdida para best estimate. Vender al actuario/CFO: "cumplan NIIF 17 con distribuciones formales, no con Excel".

**Fase 5**: Regionalizar — Mexico (CNSF), Brasil (SUSEP), Chile (CMF). Misma arquitectura, diferentes tablas y regulacion.

---

## 12) Por que 85% y no 100%

| Gap | Que es | Como se resuelve |
|---|---|---|
| Modelos actuariales complejos (GLM, credibilidad) | Tarificacion avanzada que combina multiples variables con interacciones | Parcialmente cubierto por Bayesian + Stats. GLM completo es scope de R/Python, no de motor de inferencia |
| Triangulos de desarrollo | Metodo estandar para estimar reservas de siniestros | Utility especifico (~200 lineas) o integracion con herramienta actuarial |
| Integracion RUNT / FASECOLDA | Datos vehiculares y sectoriales colombianos | Infra, no inferencia |
| Inspeccion / peritaje | Fotos, documentos, evaluacion presencial | OCR/CV externo, no inferencia |
| Core de seguros (emision, poliza, endoso) | Operacion de la poliza post-suscripcion | Integracion con Guidewire/SAP/custom, no scope |

El gap actuarial (GLM) es el mas relevante: las aseguradoras grandes usan modelos de tarificacion con GLM en R/SAS. UnderwriteIQ no reemplaza eso — consume el output del GLM como un input mas, igual que LendingBrain consume el score ML. El motor razona SOBRE el resultado actuarial, no lo reemplaza.
