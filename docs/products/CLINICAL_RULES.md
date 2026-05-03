# ClinicalRules — Motor de Soporte a Decision Clinica

Vertical: HealthTech / Clinical Decision Support. Encaje con yarumo: 80%. Score: ★★★☆☆.

Referencia: `docs/DOMAIN_ANALYSIS.md` seccion 6c (HealthTech) y seccion 7.

---

## 1) El problema

Un hospital o EPS en Colombia hoy opera con:

- **Guias de practica clinica en PDF** que el medico deberia seguir pero que nadie verifica si sigue — cada medico aplica su criterio, su experiencia, su sesgo
- **Triage subjetivo** — dos enfermeras clasifican al mismo paciente diferente segun su turno, su carga, su experiencia. El Manchester Triage System existe pero se aplica a criterio humano
- **Diagnostico diferencial en la cabeza** — el medico tiene 3-5 hipotesis, las evalua mentalmente, no queda registro de por que descarto las otras
- **Alertas de interaccion medicamentosa en el mejor caso** — si el hospital tiene HIS moderno. En la mayoria, el medico revisa mentalmente (o no revisa)
- **Cero trazabilidad clinica de la decision** — cuando Supersalud pregunta "por que le dieron de alta tan pronto?" o un comite de mortalidad investiga "se siguio el protocolo?", la respuesta es revisar la historia clinica narrativa y reconstruir
- **Glosas y rechazos** — la EPS rechaza cuentas porque el prestador no justifica formalmente el procedimiento. El prestador dice "era clinicamente necesario" pero no tiene evidencia estructurada
- **Tiempos de espera** — paciente espera 45 minutos para que el medico decida algo que un protocolo formal resolveria en segundos

Cuando un juez de tutela pregunta "por que no autorizaron este procedimiento?" o un comite de etica investiga un evento adverso, la reconstruccion de la decision clinica es arqueologia narrativa.

---

## 2) Que hace

ClinicalRules formaliza protocolos clinicos como reglas ejecutables, asiste al medico con diagnostico diferencial probabilistico, clasifica urgencia con criterios consistentes, alerta sobre interacciones y contraindicaciones, y genera trazabilidad completa de cada decision clinica. **NO reemplaza al medico — lo asiste con razonamiento formal y trazabilidad.**

**Posicionamiento regulatorio critico**: ClinicalRules es un **sistema de soporte a decision clinica (CDSS)**, NO un dispositivo medico de diagnostico. El medico siempre tiene la ultima palabra. El sistema sugiere, explica, y registra — no ordena.

---

## 3) Como usa cada paradigma

| Paradigma | Pregunta que responde | Ejemplo concreto |
|---|---|---|
| **Deductive** | "Que protocolo aplica?" | "Si fiebre > 38.5 AND tos > 7 dias AND rx torax anormal → protocolo neumonia adquirida en comunidad" |
| **Bayesian** | "Cual es el diagnostico mas probable?" | P(neumonia \| fiebre, tos, rx, leucocitos, edad, comorbilidades) vs P(TB) vs P(bronquitis) — diagnostico diferencial cuantificado |
| **Fuzzy** | "Que tan urgente es este paciente?" | "urgencia es CRITICA cuando dolor SEVERO y signos vitales INESTABLES y deterioro RAPIDO" — el triage no es binario |
| **Causal** | "Que pasa si cambiamos el protocolo?" | "Si bajamos el umbral de hospitalizacion: +15% ingresos, +$X costo, -2.3% mortalidad evitable" |
| **MCDM** | "Cual tratamiento elegir?" | Ranking dados: eficacia, efectos secundarios, costo, adherencia esperada, interacciones con medicamentos actuales, disponibilidad |
| **Temporal** | "Se cumplio el tiempo?" | "Si no mejora EN 48 horas → escalar a especialista. Si antibiotico EN > 1 hora desde ingreso a urgencias → fuera de protocolo sepsis" |
| **Predicados** | "Se completo el protocolo?" | `PARA TODO paso EN protocolo_sepsis: paso.ejecutado = true AND paso.dentro_de_tiempo = true` |
| **Explain/Audit** | "Por que esta decision?" | "Diagnostico: neumonia comunitaria (P=0.73). Protocolo: GPC neumonia adulto. Tratamiento: ampicilina/sulbactam (MCDM rank 1: eficacia 0.85, sin interacciones, disponible en formulario). Alternativa descartada: claritromicina (interaccion con warfarina del paciente)" |
| **Stats** | "Cual es la distribucion esperada?" | Poisson para tasas de eventos adversos. Normal para valores de laboratorio. Exponential para tiempo hasta respuesta al tratamiento |

---

## 4) Areas clinicas que cubre

| Area | Decisiones clave | Paradigmas dominantes |
|---|---|---|
| **Urgencias / Triage** | Clasificacion de urgencia, protocolo inicial | Fuzzy (urgencia) + Deductive (protocolo) + Temporal (tiempos) |
| **Diagnostico diferencial** | Cual es el diagnostico mas probable? Que examenes pedir? | Bayesian + Deductive (criterios diagnosticos) |
| **Prescripcion** | Cual medicamento? Interacciones? Contraindicaciones? | MCDM + Deductive (formulario, alergias) + Predicados (verificacion) |
| **Hospitalizacion vs ambulatorio** | Se hospitaliza o manejo ambulatorio? | Deductive (criterios) + Fuzzy (severidad) + Bayesian (riesgo de complicacion) |
| **Alta hospitalaria** | Es seguro dar de alta? Que seguimiento? | Predicados (criterios de alta) + Temporal (tiempo minimo) + Explain |
| **Referencia / contrarreferencia** | A que especialista? Con que urgencia? | Deductive + Fuzzy (urgencia) + Temporal (tiempos de referencia) |
| **Autorizacion de procedimientos** | Esta indicado? Cumple criterios? | Deductive + Bayesian + Explain (justificacion para EPS) |
| **Seguimiento de cronicos** | Paciente controlado? Requiere ajuste? | Temporal (controles periodicos) + Bayesian (riesgo) + WindowedStats (tendencia de valores) |
| **Farmacovigilancia** | Evento adverso? Relacionado con medicamento? | Bayesian (causalidad) + Temporal (secuencia temporal) |
| **Comite de mortalidad/morbilidad** | Se siguio el protocolo? Que fallo? | Audit/ (reconstruccion) + Explain (decisiones del caso) |

---

## 5) Flujo tipico: Urgencias

```
1. Paciente llega a urgencias

2. Enfermera registra signos vitales + motivo de consulta

3. Motor evalua TRIAGE:

   Fuzzy       → Clasificacion de urgencia (I-V Manchester adaptado)
                  dolor x signos vitales x deterioro x edad x comorbilidades
       ↓
   Temporal    → Tiempo maximo de espera segun clasificacion
       ↓
   Deductive   → Protocolo inicial recomendado segun clasificacion + motivo

4. Medico atiende, registra hallazgos

5. Motor evalua DIAGNOSTICO:

   Bayesian    → Diagnostico diferencial probabilistico
                  P(dx1 | hallazgos) = 0.73
                  P(dx2 | hallazgos) = 0.15
                  P(dx3 | hallazgos) = 0.08
       ↓
   Deductive   → Examenes sugeridos para confirmar/descartar
                  "Si P(dx1) > 0.5 AND P(dx2) > 0.1 → pedir: hemograma, PCR, rx torax"
       ↓
   Bayesian    → Actualizar con resultados de examenes (posterior)

6. Motor evalua TRATAMIENTO:

   Deductive   → Verificar alergias, contraindicaciones, interacciones
       ↓
   MCDM        → Ranking de opciones terapeuticas
                  (eficacia, efectos adversos, costo, adherencia, disponibilidad)
       ↓
   Temporal    → Tiempo para inicio de tratamiento segun protocolo
                  "Sepsis: antibiotico dentro de 1 hora"

7. Motor evalua DISPOSICION:

   Deductive   → Cumple criterios de hospitalizacion?
   Fuzzy       → Severidad integral
   Bayesian    → Riesgo de complicacion ambulatoria

8. Output (en cada paso):
   - Sugerencia clinica (NO orden — el medico decide)
   - Explicacion: por que esta sugerencia, que protocolo, que evidencia
   - Alternativas consideradas y por que se descartaron
   - Alertas: interacciones, contraindicaciones, tiempos fuera de protocolo
   - Audit trail: cada decision del medico + sugerencia del sistema + justificacion
```

---

## 6) El diferenciador: el gemelo formal del razonamiento clinico

Lo que hace ClinicalRules no es automatizar al medico — es **hacer explicito lo que el buen medico ya hace mentalmente**:

```
"Paciente M/58 anos, dolor toracico, diaforesis, disnea:

 1. Triage fuzzy: URGENCIA I (0.92 — CRITICA)
    - Dolor: SEVERO (0.9) — opresivo, irradiado a brazo izquierdo
    - Signos vitales: INESTABLES (0.7) — taquicardia 110, TA 90/60
    - Deterioro: RAPIDO (0.85) — inicio hace 30 minutos, empeorando
    → Tiempo maximo de atencion: INMEDIATO

 2. Diagnostico diferencial bayesian:
    - P(SCA/IAM)          = 0.78 — dolor tipico, diaforesis, edad, masculino
    - P(diseccion aortica) = 0.08 — TA asimetrica no reportada, pero cuadro compatible
    - P(TEP)              = 0.06 — disnea presente, sin factores de riesgo claros
    - P(neumotorax)       = 0.03 — auscultacion simetrica
    → Examenes urgentes: troponina, ECG 12 derivaciones, rx torax, dimero-D

 3. Protocolo activado: Codigo Infarto (GPC sindrome coronario agudo)
    - Aspirina 300mg VO — AHORA (verificado: sin alergia, sin contraindicacion)
    - ECG — dentro de 10 minutos desde llegada
    - Troponina — resultado en < 60 min
    - Heparina — segun resultado ECG

 4. Alertas:
    ! Paciente toma warfarina (anticoagulante) — verificar INR antes de heparina
    ! Alergia registrada a clopidogrel — alternativa: ticagrelor

 5. Causal: si los sintomas hubieran iniciado hace > 12 horas,
    P(beneficio de reperfusion) baja de 0.85 a 0.35 → cambio de protocolo"
```

Eso es lo que el cardiologo experimentado piensa en 30 segundos. ClinicalRules lo hace explicito, reproducible, y auditable. Y lo hace tambien a las 3am cuando el medico de turno tiene 2 anos de experiencia.

---

## 7) Mercado

### Quien paga

| Segmento | Tamano CO | Volumen de decisiones | Dolor |
|---|---|---|---|
| Hospitales nivel III-IV | ~80 | Miles/dia | Eventos adversos, protocolos, glosas |
| Hospitales nivel II | ~300 | Cientos/dia | Referencia/contrarreferencia, calidad |
| Clinicas privadas | ~200 | Cientos-miles/dia | Eficiencia, diferenciacion, glosas |
| EPS (aseguradoras salud) | ~30 | Autorizaciones miles/dia | Justificacion de negaciones, tutelas |
| IPS ambulatorias | ~10K | Decenas/dia cada una | Cronicos, seguimiento, adherencia |
| Redes integradas | Emergente | Variable | Coordinacion entre niveles |

### Numeros

- Gasto en salud Colombia: ~$80B COP/ano
- Glosas y rechazos: ~15-20% de facturacion hospitalaria — billones en disputa
- Tutelas en salud: ~200,000/ano (1/3 del total de tutelas del pais)
- Eventos adversos prevenibles: estimacion OMS 10% de hospitalizaciones
- Cada evento adverso evitado = vidas + millones en costos de no calidad

**La venta para hospitales**: "Reduzca eventos adversos, reduzca glosas, demuestre adherencia a protocolos."

**La venta para EPS**: "Justifique cada autorizacion/negacion formalmente. Reduzca tutelas."

---

## 8) Competencia

| Competidor | Que hace | Debilidad vs ClinicalRules |
|---|---|---|
| Epic/Cerner (CDS modules) | CDSS integrado en HIS | $$$$$$, solo para hospitales que ya tienen Epic/Cerner (casi ninguno en CO) |
| UpToDate (Wolters Kluwer) | Base de conocimiento clinico | Referencia pasiva (el medico busca), no motor de decision activo |
| Isabel Healthcare | Diagnostico diferencial | Solo dx diferencial, no multi-paradigma, no protocolos, no explicabilidad formal |
| DXplain (MIT) | Diagnostico diferencial academico | Academico, no productivo, no integrable |
| HIS colombianos (Servinte, HOSVITAL, Dinamica) | Historia clinica electronica | Son HIS, no CDSS. Almacenan datos, no razonan sobre ellos |
| Desarrollo interno | "El jefe de urgencias hizo un Excel" | No escalable, no auditable, bus factor = 1 |

**Diferenciador**: en Colombia no existe un CDSS con razonamiento multi-paradigma. Los HIS locales (Servinte, HOSVITAL) almacenan datos pero no razonan. Los CDSS internacionales (Epic CDS, Isabel) son o inaccesibles o limitados a un solo tipo de razonamiento. ClinicalRules es el motor que se integra con el HIS existente y agrega razonamiento formal.

---

## 9) Regulacion

| Norma | Que exige | Como ClinicalRules cumple |
|---|---|---|
| Resolucion 3100/2019 (Habilitacion) | Estandares de calidad, protocolos obligatorios | Deductive (protocolos formalizados) + Audit (evidencia de adherencia) |
| Decreto 780/2016 (SGSS) | Prestacion con calidad | Todos los paradigmas — decisiones formales y trazables |
| Ley 1751/2015 (Ley Estatutaria Salud) | Derecho fundamental a la salud, autonomia medica | ClinicalRules SUGIERE, no ordena. Explain/ documenta la decision del medico |
| PAMEC (Mejoramiento continuo) | Indicadores de calidad, eventos adversos | Stats + Audit (indicadores automaticos, trazabilidad de eventos) |
| Resolucion 256/2016 (IAMI) | Indicadores de calidad obligatorios | Stats (calculo automatico) + Temporal (oportunidad de atencion) |
| GPC del MinSalud | Guias de practica clinica basadas en evidencia | Deductive (GPC formalizadas como reglas ejecutables) |
| RIPS | Registro individual de prestaciones | Audit/ genera registro estructurado de cada prestacion |
| Ley 1438/2011 | Historia clinica interoperable | API estandar, integrable con HIS existente |

**Ley 1751 es la clave regulatoria**: establece la salud como derecho fundamental Y protege la autonomia medica. ClinicalRules debe posicionarse como SOPORTE, nunca como reemplazo. El medico siempre tiene la ultima palabra. El sistema hace visible el razonamiento, no lo impone.

---

## 10) Regulacion de dispositivos medicos — la precaucion

| Clasificacion INVIMA | Que es | ClinicalRules aplica? |
|---|---|---|
| Clase I | Bajo riesgo, no invasivo | Posiblemente — depende del alcance |
| Clase IIa | Riesgo moderado | Posiblemente — si el sistema influye en decisiones clinicas |
| Software como dispositivo medico (SaMD) | Software que por si solo es dispositivo medico | **GRIS** — depende de como se posicione |

**Estrategia regulatoria**: posicionar como "sistema de informacion con soporte a decision", NO como "dispositivo medico de diagnostico". La diferencia:

- **Dispositivo medico**: "El sistema diagnostica neumonia" → requiere registro INVIMA, ensayos clinicos, proceso largo
- **Sistema de soporte**: "El sistema sugiere considerar neumonia y muestra por que, el medico decide" → sistema de informacion, no dispositivo

Esta distincion es real y reconocida internacionalmente (FDA 21st Century Cures Act excluye ciertos CDSS de regulacion de dispositivos). Pero requiere cuidado en el posicionamiento y la documentacion.

---

## 11) Stack tecnico

```
┌──────────────────────────────────────────────────┐
│  Frontend (integrado en HIS o standalone)          │
├──────────────────────────────────────────────────┤
│  API REST/gRPC + HL7 FHIR adapter                 │
│  ┌────────────────────────────────────────┐       │
│  │ Orchestrator (flujo clinico)           │       │
│  │ Triage → Dx → Examenes → Tratamiento → │       │
│  │ Disposicion → Seguimiento               │       │
│  └────────────────────────────────────────┘       │
├──────────────────────────────────────────────────┤
│  SDK decision-engine/core/                        │
│  (decisions + explain + validate + audit)         │
├──────────────────────────────────────────────────┤
│  yarumo compute/ + maths/                       │
│  (deductive, bayesian, fuzzy, causal, mcdm)      │
├──────────────────────────────────────────────────┤
│  Base de conocimiento clinico                     │
│  ┌────────────┬───────────┬───────────────────┐  │
│  │ GPC MinSalud│ Vademecum │ Interacciones    │  │
│  │ (reglas)    │ (farmacos)│ (contraindicac.) │  │
│  │ CIE-10/11  │ CUPS      │ Valores ref lab  │  │
│  └────────────┴───────────┴───────────────────┘  │
├──────────────────────────────────────────────────┤
│  Integraciones                                    │
│  ┌────────────┬───────────┬───────────────────┐  │
│  │ HIS        │ LIS       │ RIS/PACS          │  │
│  │(Servinte,  │(laborat.) │(imagenes)         │  │
│  │ HOSVITAL)  │           │                   │  │
│  └────────────┴───────────┴───────────────────┘  │
├──────────────────────────────────────────────────┤
│  Storage + Telemetry                              │
└──────────────────────────────────────────────────┘
```

**HL7 FHIR** es critico: es el estandar de interoperabilidad en salud. Sin un adapter FHIR, ClinicalRules no se integra con HIS modernos. Con FHIR, se integra con cualquiera.

---

## 12) Go-to-market

**Fase 1**: Un hospital nivel III como design partner. Caso: urgencias — triage fuzzy + protocolo sepsis (temporal: antibiotico en < 1 hora) + diagnostico diferencial bayesian para dolor toracico. Demostrar reduccion en tiempo de clasificacion y adherencia medible a protocolos.

**Fase 2**: Agregar prescripcion asistida — verificacion de interacciones + MCDM para seleccion de medicamento. Demostrar reduccion de eventos adversos por medicamentos.

**Fase 3**: Modulo de justificacion para EPS — cada procedimiento/medicamento con explicacion formal de indicacion clinica. Reducir glosas y argumentar tutelas.

**Fase 4**: Expandir a cronicos — seguimiento de diabeticos, hipertensos, EPOC. WindowedStats para tendencia de HbA1c, creatinina, espirometria. Temporal para controles periodicos.

**Fase 5**: Base de conocimiento clinico como servicio — GPC del MinSalud formalizadas como reglas ejecutables. Vender/licenciar la base de conocimiento, no solo el motor.

---

## 13) Por que 80% y no 100%

| Gap | Que es | Como se resuelve | Impacto |
|---|---|---|---|
| **Interoperabilidad HL7/FHIR** | Estandar de datos clinicos | Adapter FHIR (~500-800 lineas) — no es inferencia pero es bloqueante para integracion | **Alto** |
| **Base de conocimiento clinico** | GPC formalizadas, vademecum, interacciones, CIE-10 | Trabajo de contenido (medicos + ingenieros), no de motor | **Alto** |
| **Regulacion INVIMA/SaMD** | Clasificacion como dispositivo medico o no | Proceso legal/regulatorio, no tecnico | **Medio** |
| **Imagenes diagnosticas** | Rx, TAC, RM — interpretacion | CV/AI externo, completamente fuera de scope | **Bajo** |
| **Senales de monitoreo continuo** | ECG, oximetria, presion continua | IoT stream processing, parcialmente cubierto con WindowedStats | **Bajo** |
| **NLP clinico** | Extraer datos estructurados de notas medicas narrativas | LLM companion, no inferencia | **Medio** |

Los 2 gaps altos (FHIR + base de conocimiento) son los que bajan el score a ★★★☆☆. No son limitaciones del motor de inferencia — son **barreras de entrada al dominio salud**:

1. **FHIR** requiere un adapter especifico (no trivial pero no enorme)
2. **Base de conocimiento** requiere medicos que formalicen GPC como reglas — es trabajo intensivo en dominio, no en codigo

El motor hace el 80% del razonamiento clinico. El 20% es acceder al dominio (datos + conocimiento + regulacion).

---

## 14) Por que ★★★☆☆ y no ★★★★☆

A pesar del 80% de encaje tecnico, el score es menor que GovRules porque:

| Factor | GovRules (★★★★☆) | ClinicalRules (★★★☆☆) |
|---|---|---|
| Barrera de entrada | Baja — reglas son texto legal publico | Alta — requiere medicos para formalizar GPC |
| Complejidad de datos | Baja — SISBEN, BDUA son estructurados | Alta — HL7/FHIR, datos clinicos heterogeneos |
| Regulacion de producto | No aplica (software de gestion) | Posible clasificacion como dispositivo medico |
| Time-to-market | 3-6 meses para MVP | 6-12 meses minimo (FHIR + base de conocimiento) |
| Riesgo | Bajo — peor caso: decision administrativa incorrecta | **Alto — peor caso: dano al paciente** |

El ultimo punto es el mas importante: en salud, un error del motor puede contribuir a dano al paciente. Esto requiere validacion clinica rigurosa, no solo tests unitarios. El motor de inferencia es correcto por diseno, pero la base de conocimiento que alimenta las reglas debe ser validada por comites medicos.
