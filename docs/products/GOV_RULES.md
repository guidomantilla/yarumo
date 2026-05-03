# GovRules — Motor de Decision para Politica Publica

Vertical: GovTech / Government Decision-Making. Encaje con yarumo: 80%. Score: ★★★★☆.

Referencia: `docs/DOMAIN_ANALYSIS.md` seccion 6f (GovTech) y seccion 7.

---

## 1) El problema

Un funcionario publico que administra programas sociales, subsidios, permisos, o fiscalizacion hoy trabaja con:

- **Resoluciones en PDF** que nadie traduce a logica ejecutable — cada funcionario interpreta la norma distinto
- **Sistemas legacy en COBOL/Oracle Forms/Java 6** que nadie puede modificar — cambiar un umbral de elegibilidad toma meses
- **Discrecionalidad invisible** — dos ciudadanos con el mismo perfil reciben respuestas diferentes segun el funcionario, la oficina, o el dia
- **Cero trazabilidad de decisiones masivas** — "por que 50,000 familias quedaron fuera del subsidio?" → nadie puede reconstruirlo
- **Focalizacion por proxy** — "pobre es el que tiene SISBEN < X" pero X se calibra politicamente, no tecnicamente
- **Auditoria reactiva** — la Contraloria llega 2 anos despues y el funcionario no puede demostrar POR QUE se tomo cada decision

Cuando un juez de tutela pregunta "por que le negaron el subsidio a esta persona?", la entidad busca en archivos fisicos. Cuando la Contraloria pregunta "como focalizaron estos $500.000M?", la respuesta es un PowerPoint.

---

## 2) Que hace

GovRules formaliza politicas publicas como reglas ejecutables, evalua elegibilidad de forma consistente, prioriza asignacion de recursos con criterios explicitos, monitorea plazos legales, y genera evidencia auditable automatica para cada decision. No reemplaza la politica publica — la hace ejecutable, verificable, y explicable.

---

## 3) Como usa cada paradigma

| Paradigma | Pregunta que responde | Ejemplo concreto |
|---|---|---|
| **Deductive** | "Esta persona es elegible?" | "Si SISBEN < 40 AND edad > 60 AND no_pensionado AND no_beneficiario_otro_programa → elegible Colombia Mayor" |
| **Bayesian** | "Cual es la probabilidad real de vulnerabilidad?" | P(vulnerable \| indicadores socioeconomicos, zona, composicion hogar) — mas alla del corte SISBEN binario |
| **Fuzzy** | "Que tan urgente es este caso?" | "urgencia es CRITICA cuando pobreza EXTREMA y acceso_salud NULO y menores_a_cargo ALTO" — la urgencia no es binaria |
| **Causal** | "Que pasa si cambiamos el umbral?" | "Si subimos el SISBEN de 40 a 45: +120K beneficiarios, +$80.000M presupuesto, de donde sale? que programas se afectan?" |
| **MCDM** | "Cual proyecto priorizar?" | Ranking dados: impacto social, costo, viabilidad tecnica, poblacion beneficiada, cobertura geografica, alineacion con PND |
| **Temporal** | "Estamos dentro de los plazos legales?" | Derecho de peticion: 15 dias habiles. Tutela: 10 dias. PQRS: segun entidad. Desembolso de subsidio: segun resolucion |
| **Predicados** | "Todos los beneficiarios cumplen los requisitos?" | `PARA TODO beneficiario EN programa: beneficiario.documentos_completos AND beneficiario.no_duplicado_en_otro_programa` |
| **Explain/Audit** | "Por que esta persona si/no?" | Trace completo: que reglas se evaluaron, con que datos, que resultado, cuando. Listo para Contraloria, Procuraduria, o juez de tutela |

---

## 4) Tipos de decision gubernamental que cubre

| Tipo de decision | Ejemplo | Paradigmas dominantes |
|---|---|---|
| **Elegibilidad** | Tiene derecho al subsidio/programa? | Deductive + Predicados |
| **Focalizacion** | A quienes llegar primero con recursos limitados? | Bayesian + Fuzzy + MCDM |
| **Priorizacion de inversion** | Cual proyecto de infraestructura ejecutar primero? | MCDM + Causal |
| **Fiscalizacion** | A quien auditar/inspeccionar? | Bayesian (riesgo) + Deductive (criterios legales) |
| **Permisos y licencias** | Se otorga la licencia ambiental/construccion? | Deductive + Predicados + Temporal |
| **Respuesta a PQRS** | Se resuelve a favor? En plazo? | Deductive + Temporal + Explain |
| **Asignacion de cupos** | Quien entra al programa? (educacion, vivienda, salud) | Deductive (elegibilidad) + Fuzzy (priorizacion) + MCDM (ranking) |
| **Calibracion de politica** | Que pasa si cambiamos el criterio X? | Causal + Stats |

---

## 5) Flujo tipico: Programa social

```
1. Ciudadano solicita beneficio (ventanilla/web/app)
   O: entidad cruza bases de datos para focalizacion activa

2. Datos entran al motor:
   - SISBEN IV (puntaje + variables desagregadas)
   - BDUA/ADRES (afiliacion salud)
   - Registraduria (identidad, edad)
   - DANE (zona, estrato, NBI)
   - Bases propias (historial de beneficios, cumplimiento de compromisos)
   - Otros programas (cruce anti-duplicacion)

3. Motor evalua:

   Predicados  → Documentacion completa? No duplicado en otro programa?
       ↓ si completo
   Deductive   → Cumple criterios de elegibilidad? (resolucion vigente)
       ↓ si elegible
   Bayesian    → Score de vulnerabilidad real (mas alla del corte SISBEN)
       ↓
   Fuzzy       → Urgencia integral (pobreza x acceso x composicion hogar)
       ↓
   MCDM        → Ranking entre elegibles cuando hay mas demanda que cupos
       ↓
   Temporal    → Plazo de respuesta al ciudadano? Plazo de desembolso?

4. Output:
   - ELEGIBLE / NO ELEGIBLE / LISTA DE ESPERA (con posicion)
   - Prioridad relativa vs otros elegibles
   - Explicacion completa: por que si/no, que criterios aplicaron
   - Si no elegible: que le falta y a que otros programas podria aplicar
   - Audit trail: para Contraloria, Procuraduria, juez de tutela
   - Estadisticas agregadas: cuantos elegibles, distribucion geografica, presupuesto requerido
```

---

## 6) El diferenciador: transparencia formal

Lo que hace unico a GovRules no es automatizar — es **hacer visible la logica de la decision publica**:

```
"Solicitud Colombia Mayor — Maria Garcia, CC 51.234.567:

 1. Documentacion: COMPLETA (cedula ✓, SISBEN ✓, declaracion ✓)
 2. Elegibilidad: CUMPLE
    - Edad: 68 anos (>= 65 ✓)
    - SISBEN IV: 32.4 (< 40 ✓)
    - No pensionada ✓
    - No beneficiaria Familias en Accion ✓
 3. Vulnerabilidad bayesian: 0.78 (alta)
    - Zona rural dispersa (+0.12)
    - Hogar unipersonal (+0.08)
    - Sin ingresos formales (+0.15)
 4. Urgencia fuzzy: ALTA (0.81)
    - Pobreza: ALTA (0.75)
    - Acceso salud: BAJO (0.68) — EPS subsidiada, IPS a 2h
    - Red de apoyo: BAJA (0.85) — sin familiares en municipio
 5. Ranking MCDM: posicion 342 de 15,820 elegibles en el departamento
    - Criterios: vulnerabilidad (40%), urgencia (30%), antiguedad solicitud (20%), zona (10%)

 Decision: ELEGIBLE — PRIORIDAD ALTA
 Plazo de respuesta: 15 dias habiles (vence 2026-03-18)
 Plazo de inclusion en nomina: siguiente ciclo de pagos"
```

Eso responde la tutela. Eso responde la Contraloria. Eso responde la veeduria ciudadana. Y lo genera automaticamente para cada uno de los 50,000 beneficiarios, no solo para el que puso la tutela.

---

## 7) Programas sociales colombianos que cubre naturalmente

| Programa | Entidad | Decision | Paradigmas |
|---|---|---|---|
| **Colombia Mayor** | MinTrabajo/Consorcio | Elegibilidad + priorizacion de adultos mayores | Deductive + Fuzzy + MCDM |
| **Familias en Accion** | DPS | Elegibilidad + verificacion de compromisos (asistencia escolar, controles salud) | Deductive + Predicados + Temporal |
| **Jovenes en Accion** | DPS | Elegibilidad + permanencia en programa educativo | Deductive + Temporal + Predicados |
| **Ingreso Solidario** | DNP/DPS | Focalizacion + cruce de bases + anti-duplicacion | Deductive + Predicados + Bayesian |
| **Mi Casa Ya** | MinVivienda/Fonvivienda | Elegibilidad + priorizacion de subsidio de vivienda | Deductive + MCDM + Temporal |
| **ICETEX** | ICETEX | Elegibilidad credito educativo + priorizacion | Deductive + Bayesian + MCDM |
| **Victimas (Ley 1448)** | UARIV | Elegibilidad + medidas de reparacion + plazos | Deductive + Temporal + Fuzzy + Explain |
| **Licencias ambientales** | ANLA/CARs | Evaluacion de requisitos + condiciones + seguimiento | Deductive + Predicados + Temporal |
| **Fiscalizacion tributaria** | DIAN | Seleccion de contribuyentes a auditar | Bayesian (riesgo) + Deductive (criterios) |
| **PQRS** | Cualquier entidad | Respuesta en plazo + resolucion fundamentada | Temporal + Deductive + Explain |

---

## 8) Mercado

### Quien paga

| Segmento | Tamano CO | Dolor | Presupuesto |
|---|---|---|---|
| Entidades del orden nacional | ~200 (ministerios, departamentos admin, agencias) | Auditoria CGR, tutelas, transparencia | Alto (presupuesto publico) |
| Gobernaciones | 32 | Mismos dolores + rendicion de cuentas local | Medio-Alto |
| Alcaldias capitales | 32 | Alto volumen de PQRS, programas sociales propios | Medio-Alto |
| Alcaldias municipales | ~1,100 | Baja capacidad tecnica, alta necesidad | Bajo individual, alto agregado |
| Superintendencias | ~12 | Decisiones sancionatorias explicables | Alto |
| Organismos de control | CGR, PGN, Defensoria | Herramientas para VERIFICAR decisiones de otros | Alto |
| DNP / DANE | 2 | Diseno y evaluacion de politica publica | Alto |

### Numeros

- Presupuesto General de la Nacion 2026: ~$500B COP
- Programas sociales: ~$30B COP/ano solo en transferencias directas
- Tutelas contra entidades publicas: ~600,000/ano
- Derechos de peticion: millones/ano
- Costo de una tutela perdida: no solo dinero — es precedente judicial + desgaste institucional

**La venta no es software — es defensa institucional**: "Cuando el juez pregunte por que, usted tiene la respuesta. Cuando la Contraloria audite, usted tiene la evidencia. Cuando la veeduria cuestione, usted tiene la transparencia."

---

## 9) Competencia

| Competidor | Que hace | Debilidad vs GovRules |
|---|---|---|
| SINERGIA (DNP) | Seguimiento a PND, indicadores | Monitoreo de indicadores, no motor de decision |
| SUIT (DAFP) | Gestion de tramites | Catalogo de tramites, no decision |
| Sistemas propios (cada entidad) | Desarrollos custom en Java/.NET | Silos, no auditables, no explicables, no reutilizables |
| BPM (Bizagi, Bonita) | Workflow/flujos de proceso | Automatizan el PROCESO, no la DECISION dentro del proceso |
| Salesforce Government Cloud | CRM gubernamental | CRM, no razonamiento formal |
| IBM ODM (caso gobierno) | Motor de reglas enterprise | $$$$, requiere consultora, overkill para la mayoria de entidades |

**Diferenciador**: no existe en Colombia (ni en LATAM) un motor de decision gubernamental con razonamiento multi-paradigma y explicabilidad formal. Lo mas cercano es un desarrollo custom por entidad — caro, no reutilizable, no auditable. GovRules es el motor comun que cada entidad configura con sus reglas.

---

## 10) Regulacion que lo hace obligatorio

| Norma | Que exige | Como GovRules cumple |
|---|---|---|
| Ley 1712/2014 (Transparencia) | Publicar criterios de decision, informacion activa | Las reglas SON los criterios, Explain/ los hace publicos |
| Ley 1755/2015 (PQRS) | Respuesta fundamentada en plazo | Temporal (plazos) + Explain (fundamentacion) |
| Decreto 2591/1991 (Tutela) | Fallo en 10 dias, motivado | Temporal + Explain + Audit |
| Ley 1437/2011 (CPACA) | Actos administrativos motivados | Explain/ genera motivacion formal automatica |
| CONPES 3956 (Focalizacion) | Focalizacion objetiva, verificable | Deductive + Bayesian + MCDM — criterios explicitos |
| MIPG (Modelo Integrado) | Gestion por resultados, rendicion de cuentas | Audit/ + Stats (indicadores de gestion) |
| Ley 2052/2020 (Gobierno Digital) | Servicios digitales, interoperabilidad | API-first, integrable con GOV.CO |
| CGR (auditoria fiscal) | Demostrar uso eficiente de recursos publicos | Audit trail completo, MCDM para priorizacion con criterios explicitos |

**CPACA (Ley 1437) es la clave**: todo acto administrativo debe ser MOTIVADO. Es decir, la entidad debe explicar POR QUE tomo la decision, con fundamento en hechos y normas. GovRules genera esa motivacion automaticamente. Hoy los funcionarios escriben motivaciones a mano (o con copy-paste) para miles de resoluciones.

---

## 11) Stack tecnico

```
┌──────────────────────────────────────────────────┐
│  Frontend (portal funcionario + portal ciudadano) │
├──────────────────────────────────────────────────┤
│  API REST/gRPC                                    │
│  ┌────────────────────────────────────────┐       │
│  │ Orchestrator (cascada de decision)     │       │
│  │ Docs → Elegibilidad → Vulnerabilidad → │       │
│  │ Priorizacion → Plazos → Motivacion     │       │
│  └────────────────────────────────────────┘       │
├──────────────────────────────────────────────────┤
│  SDK decision-engine/core/                        │
│  (decisions + explain + validate + audit)         │
├──────────────────────────────────────────────────┤
│  yarumo compute/ + maths/                       │
│  (deductive, bayesian, fuzzy, causal, mcdm)      │
├──────────────────────────────────────────────────┤
│  Integraciones externas                           │
│  ┌────────────┬────────────┬──────────────────┐  │
│  │ SISBEN IV  │ BDUA/ADRES │ Registraduria    │  │
│  │ DANE       │ DIAN       │ Otros programas  │  │
│  │ (X-Road CO)│            │ (anti-duplicacion)│  │
│  └────────────┴────────────┴──────────────────┘  │
├──────────────────────────────────────────────────┤
│  Storage + Telemetry + NLP companion              │
└──────────────────────────────────────────────────┘
```

---

## 12) Go-to-market

**Fase 1**: Una entidad del orden nacional como design partner. Candidatos ideales: DPS (Familias en Accion — alto volumen, reglas claras, tutelas frecuentes) o UARIV (Victimas — alta presion judicial, reglas complejas, necesidad urgente de explicabilidad). Implementar elegibilidad + plazos + explicacion automatica de decisiones.

**Fase 2**: Agregar focalizacion bayesian + priorizacion MCDM. Demostrar que la focalizacion es mas justa y verificable que el corte SISBEN binario.

**Fase 3**: Expandir a 3-5 entidades nacionales. Agregar causal "what-if" para DNP/planeacion: "si cambiamos el umbral de elegibilidad, cuantos beneficiarios nuevos, cuanto presupuesto adicional, que departamentos se afectan mas".

**Fase 4**: Alcaldias capitales — Bogota, Medellin, Cali, Barranquilla. Programas sociales locales + PQRS masivos.

**Fase 5**: Modelo SaaS multi-tenant para municipios medianos. Subsidiado por MinTIC/gobierno digital o cooperacion internacional (BID, Banco Mundial — ambos financian modernizacion del Estado en LATAM).

---

## 13) La jugada estrategica: cooperacion internacional

El BID, Banco Mundial, y CAF financian proyectos de modernizacion estatal en LATAM. "Motor de decision transparente para politica publica" es exactamente el tipo de proyecto que financian — governance, transparencia, eficiencia del gasto publico, reduccion de discrecionalidad.

El pitch no es "compreme software". Es "financien la implementacion de transparencia algoritmica en decisiones de politica social". GovRules es la herramienta, la cooperacion es el canal.

---

## 14) Por que 80% y no 100%

| Gap | Que es | Como se resuelve |
|---|---|---|
| Interoperabilidad X-Road/GOV.CO | Conectar con bases de datos del Estado (SISBEN, BDUA, Registraduria) | Infra — Colombia tiene X-Road desde 2020, pero la adopcion es desigual |
| Volumen masivo | Algunos programas tienen millones de beneficiarios | Arquitectura — el motor es rapido por diseno (Go), pero el orquestador necesita batch processing |
| Gestion documental | Resoluciones, actas, soportes | Fuera de scope — integracion con gestion documental existente (Orfeo, SGDEA) |
| Firma electronica | Actos administrativos requieren firma del funcionario | Integracion con Certicamara/firma digital, no inferencia |
| Calendario habil gubernamental | Dias habiles varian por entidad, incluye puentes festivos (Ley Emiliani) | Utility pequeno o integracion con calendario institucional |

El 20% faltante es plomeria institucional — interoperabilidad con sistemas del Estado, firma electronica, gestion documental. Ninguno es inferencia. El motor hace el trabajo intelectual de la decision; el 20% es conectarlo con la burocracia existente.
