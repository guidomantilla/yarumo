# AgroDecide — Motor de Decision para Agricultura de Precision

Vertical: AgriTech / Smart Agriculture. Encaje con yarumo: 75%. Score: ★★★☆☆.

Referencia: `docs/DOMAIN_ANALYSIS.md` seccion 6e (AgriTech) y seccion 7.

---

## 1) El problema

Un agricultor o agronomo en Colombia hoy opera con:

- **Intuicion y tradicion** — "mi abuelo sembraba en octubre" funciona hasta que el clima cambia. Y el clima ya cambio
- **Asistencia tecnica esporadica** — el extensionista de la federacion/UMATA visita cada 3-6 meses. Entre visitas, el agricultor decide solo
- **Alertas genericas** — "hay riesgo de roya en la zona" pero no "en TU finca, con TU variedad, dado TU microclima, el riesgo es X"
- **Sensores sin cerebro** — quien invierte en IoT tiene datos de humedad, temperatura, pH, pero no tiene QUIEN razone sobre esos datos para tomar decisiones
- **Decisiones de cultivo a ciegas** — "siembro cafe o aguacate?" basado en lo que el vecino sembro, no en analisis multi-criterio de suelo, clima, mercado, y agua
- **Cero trazabilidad para certificacion** — GlobalGAP, Rainforest Alliance, organicos, Fair Trade exigen demostrar decisiones de manejo. El registro es un cuaderno
- **Perdidas evitables** — Colombia pierde ~30% de produccion agricola por plagas, enfermedades, y manejo inadecuado (FAO). Cafe: broca y roya pueden devastar hasta 40% de la cosecha

Cuando la federacion de cafeteros pregunta "por que no fumigo a tiempo?" o el certificador pregunta "que criterio uso para aplicar el fungicida?", la respuesta es "no me di cuenta" o "asi se hace aqui".

---

## 2) Que hace

AgroDecide toma datos del campo (sensores IoT, observaciones del agricultor, datos climaticos), aplica reglas agronomicas formalizadas, evalua riesgo de plagas/enfermedades, prioriza acciones, y genera recomendaciones explicables y trazables. No reemplaza al agronomo — extiende su conocimiento a miles de fincas simultaneamente.

---

## 3) Como usa cada paradigma

| Paradigma | Pregunta que responde | Ejemplo concreto |
|---|---|---|
| **Deductive** | "Que accion tomar dado estas condiciones?" | "Si humedad_suelo < 30% AND pronostico_lluvia = 0 AND etapa = floracion → regar AHORA. Si temperatura < 5°C AND cultivo = cafe → activar protocolo helada" |
| **Bayesian** | "Cual es la probabilidad de plaga/enfermedad?" | P(roya \| humedad relativa > 80%, temperatura 18-25°C, variedad susceptible, historial finca, prevalencia zona) |
| **Fuzzy** | "Cual es el estado general del cultivo?" | "salud es REGULAR cuando color AMARILLENTO y crecimiento LENTO y hojas MANCHADAS" — el estado del cultivo no es binario |
| **Causal** | "Que pasa si cambio la practica?" | "Si cambio de variedad Caturra a Castillo: -85% susceptibilidad roya, -5% taza (flavor), +2 anos hasta produccion plena" |
| **MCDM** | "Que cultivo/variedad sembrar?" | Ranking dados: rendimiento esperado, resistencia a plagas, requerimiento hidrico, precio de mercado, costo de insumos, certificabilidad |
| **Temporal** | "Es el momento correcto?" | "Si floracion detectada AND ultimos 45 dias sin fertilizacion → fertilizar AHORA. Si 3 dias consecutivos HR > 85% → alerta preventiva roya" |
| **Predicados** | "Se completo el plan de manejo?" | `PARA TODO lote EN finca: lote.analisis_suelo_vigente AND lote.plan_fertilizacion_aplicado` |
| **Explain/Audit** | "Por que esta recomendacion?" | "Recomendacion: aplicar fungicida preventivo lote 3. Razon: P(roya) = 0.72 (bayesian: HR promedio 87% ultimos 5 dias + variedad Caturra + finca vecina con brote). Alternativa: monitoreo intensivo si P(roya) < 0.5" |
| **Stats/WindowedStats** | "Como viene la tendencia?" | Promedio de temperatura ultimos 7 dias, varianza de humedad en ventana de 24h, tendencia de pH del suelo |

---

## 4) Cultivos colombianos que cubre naturalmente

| Cultivo | Area CO | Decisiones clave | Paradigmas dominantes |
|---|---|---|---|
| **Cafe** | ~850K hectareas, ~540K familias | Roya, broca, fertilizacion, cosecha, variedad | Bayesian (roya/broca) + Fuzzy (estado) + Temporal (ciclos) + MCDM (variedad) |
| **Flores** | ~8K hectareas, exportacion $2B USD | Control ambiental (invernadero), plagas, calidad de corte | IoT + Fuzzy + Temporal (ciclo de corte) + Deductive (protocolos) |
| **Palma de aceite** | ~580K hectareas | Pudricion de cogollo, fertilizacion, cosecha optima | Bayesian (PC) + Temporal + Fuzzy (estado racimo) |
| **Banano** | ~50K hectareas, exportacion | Sigatoka negra, Fusarium R4T (amenaza existencial), calidad exportacion | Bayesian (Fusarium) + Deductive (cuarentena) + Temporal |
| **Cana de azucar** | ~240K hectareas, Valle del Cauca | Maduracion optima, riego, plagas | Stats (sacarosa) + Temporal (cosecha) + Fuzzy |
| **Aguacate Hass** | Creciente, exportacion | Antracnosis, momento de cosecha, manejo post-cosecha | Bayesian (antracnosis) + Temporal + MCDM |
| **Cacao** | ~190K hectareas | Monilia, escoba de bruja, fermentacion | Bayesian (enfermedades) + Temporal (fermentacion) + Fuzzy |

---

## 5) Flujo tipico: Finca cafetera

```
1. Datos entran al motor (continuo):
   - Sensores IoT: temperatura, humedad relativa, humedad suelo, lluvia
   - Estacion meteorologica: pronostico 7 dias
   - Observacion del cafetero (app): "hojas amarillas lote 3", "broca en cereza"
   - Datos de zona: Cenicafe alertas, prevalencia roya departamento
   - Historial finca: variedad, edad del cafetal, aplicaciones previas

2. Motor evalua DIARIAMENTE:

   WindowedStats → Tendencia de variables ambientales (ventana 7 dias)
       ↓
   Temporal    → Condiciones sostenidas para roya/broca?
                  "HR > 80% por > 5 dias consecutivos en rango 18-25°C"
       ↓
   Bayesian    → P(roya lote 3) = 0.72
                  P(broca lote 1) = 0.45
       ↓
   Fuzzy       → Estado general: lote 1 BUENO, lote 2 BUENO, lote 3 REGULAR
       ↓
   Deductive   → Acciones segun umbrales y estado:
                  "P(roya) > 0.6 AND estado REGULAR → aplicar fungicida preventivo"
                  "P(broca) > 0.3 AND cereza madura → trampeo + re-re"
       ↓
   MCDM        → Priorizar acciones (si hay varias simultaneas):
                  "Roya lote 3 primero (urgencia + impacto economico > broca lote 1)"

3. Motor evalua ESTACIONALMENTE:

   MCDM        → Plan de renovacion: que lotes renovar? que variedad?
                  (edad x productividad x susceptibilidad x costo x mercado)
       ↓
   Causal      → "Si renuevo lote 3 con Castillo: en 3 anos, P(roya) baja de
                  0.72 a 0.08, produccion sube 15%, pierdo 2 cosechas"

4. Output:
   - Notificacion al cafetero (app/SMS): "Lote 3: aplicar fungicida esta semana.
     Razon: riesgo alto de roya (72%). Ver detalle."
   - Plan de manejo semanal priorizado
   - Registro de manejo para certificacion (GlobalGAP, Rainforest Alliance)
   - Dashboard para extensionista: estado de todas las fincas asignadas
   - Alertas tempranas: "Si no actua en 5 dias, riesgo sube a 89%"
```

---

## 6) El diferenciador: el extensionista digital

Lo que hace AgroDecide no es reemplazar al extensionista — es **multiplicarlo por 1,000**:

```
"Finca El Porvenir, vereda La Esperanza, Planadas (Tolima):

 Reporte semanal — 3 de marzo 2026

 LOTE 1 (Castillo, 3 anos, 1.2 ha):
   Estado: BUENO (fuzzy 0.82)
   Riesgos: bajo (P(roya)=0.05, variedad resistente)
   Accion: ninguna. Continuar monitoreo.

 LOTE 2 (Colombia, 6 anos, 0.8 ha):
   Estado: BUENO (fuzzy 0.75)
   Riesgos: moderado (P(broca)=0.38 — cereza madura expuesta)
   Accion: instalar 3 trampas/ha + recoleccion oportuna
   Urgencia: MEDIA — actuar esta semana

 LOTE 3 (Caturra, 12 anos, 1.5 ha):
   Estado: REGULAR (fuzzy 0.45)
   Riesgos: ALTO (P(roya)=0.72 — variedad susceptible + HR sostenida)
   Accion: aplicar fungicida cuprico preventivo
   Urgencia: ALTA — actuar en 3 dias
   Alternativa: si prefiere no aplicar quimico → poda sanitaria + monitoreo
     diario (P(roya) sube a 0.89 en 7 dias sin accion)

 RECOMENDACION ESTACIONAL:
   Lote 3 candidato a renovacion. Variedad sugerida: Castillo (MCDM rank 1:
   resistencia roya 0.95, productividad +15%, calidad taza comparable).
   Costo estimado: $2.8M/ha. Recuperacion: 3 cosechas (2029).

 Este reporte fue generado automaticamente. Consulte con su extensionista
 para confirmar las recomendaciones."
```

Eso llega al celular del cafetero en Planadas. Y llega al celular de los otros 999 cafeteros de la zona. Y al extensionista le llega un dashboard con los 1,000 fincas priorizadas: cuales necesitan visita urgente, cuales van bien.

---

## 7) Mercado

### Quien paga

| Segmento | Tamano CO | Dolor | Modelo de pago |
|---|---|---|---|
| Federacion de Cafeteros (FNC) | ~540K familias cafeteras | Roya, broca, productividad, renovacion | Institucional — FNC como cliente, servicio a cafeteros |
| Asocolflores | ~8K ha, ~130 empresas | Control ambiental, calidad exportacion | Empresa por empresa, ROI claro |
| Fedepalma | ~580K ha, ~6K productores | Pudricion de cogollo, sostenibilidad | Institucional + empresa |
| Augura/Asbama (banano) | ~50K ha | Fusarium R4T (emergencia fitosanitaria) | Gremio + gobierno (defensa sanitaria) |
| Asocana | ~240K ha | Optimizacion de cosecha, riego | Empresa (ingenios azucareros) |
| Agroexportadores (aguacate, cacao) | Creciente | Certificacion, calidad, trazabilidad | Empresa, subsidiado por programas de exportacion |
| ICA (Instituto Colombiano Agropecuario) | Nacional | Vigilancia fitosanitaria, alertas tempranas | Gobierno — herramienta de vigilancia |
| Agencias de cooperacion | BID, FAO, USAID | Productividad rural, seguridad alimentaria | Financiamiento de proyectos |

### Numeros

- PIB agropecuario Colombia: ~$50B COP/ano (~7% del PIB)
- Cafe: exportaciones ~$3.5B USD/ano — tercer exportador mundial
- Flores: exportaciones ~$2B USD/ano — segundo exportador mundial
- Perdidas por roya en cafe: hasta 30% de produccion en anos malos
- Fusarium R4T en banano: amenaza existencial — un brote podria destruir la industria bananera colombiana
- Costo de un extensionista: ~$3-5M COP/mes. AgroDecide escala sin headcount

---

## 8) Competencia

| Competidor | Que hace | Debilidad vs AgroDecide |
|---|---|---|
| Cenicafe (alertas) | Alertas de roya por zona basadas en clima | Genericas (por zona, no por finca), no son decisiones, no son explicables |
| CropIn | Plataforma AgTech global | General, no tiene razonamiento formal, mas BI que decision |
| Farmbeat (Microsoft) | IoT + ML para agricultura | ML black-box, no explicable, pricing enterprise, no LATAM-first |
| Climate FieldView (Bayer) | Datos agronomicos + recomendaciones | Sesgado a productos Bayer, no abierto, US-centric |
| Agronet (MinAgricultura CO) | Informacion agricola publica | Informacion estatica, no motor de decision |
| Extensionista humano | Visita presencial, conocimiento experto | No escala — 1 extensionista por ~500 fincas, visita cada 3-6 meses |

**Diferenciador**: los competidores son o plataformas genericas de BI agricola o alertas genericas por zona. Ninguno combina IoT + razonamiento multi-paradigma + explicabilidad + trazabilidad para certificacion. AgroDecide formaliza el conocimiento del extensionista experto y lo lleva a 1,000 fincas.

---

## 9) Regulacion y certificacion

| Norma/Certificacion | Que exige | Como AgroDecide cumple |
|---|---|---|
| GlobalGAP | Registro de aplicaciones, justificacion de uso de agroquimicos | Audit/ (registro automatico de cada decision y aplicacion) |
| Rainforest Alliance | Manejo integrado de plagas, trazabilidad | Explain/ (por que se aplico, por que no se uso alternativa organica) |
| Organico (USDA/EU) | Cero agroquimicos, manejo biologico documentado | Deductive (reglas de manejo organico) + Audit |
| Fair Trade | Practicas sostenibles documentadas | Audit/ + Explain/ |
| Resolucion ICA (fitosanitaria) | Reporte obligatorio de plagas cuarentenarias (Fusarium R4T) | Bayesian (deteccion temprana) + Temporal (reporte inmediato) + Deductive (protocolo cuarentena) |
| BPA (Buenas Practicas Agricolas) | Registro de manejo, trazabilidad | Audit/ genera registro BPA automaticamente |

**Fusarium R4T en banano es una emergencia**: Colombia declaro emergencia fitosanitaria. Un sistema que detecte tempranamente (bayesian: probabilidad dado sintomas + zona + vectores) y active protocolos de cuarentena (deductive + temporal) tiene valor incalculable — literalmente puede salvar la industria bananera de una zona.

---

## 10) Stack tecnico

```
┌──────────────────────────────────────────────────┐
│  Frontend (app movil cafetero + dashboard web)     │
├──────────────────────────────────────────────────┤
│  API REST/gRPC                                    │
│  ┌────────────────────────────────────────┐       │
│  │ Orchestrator (ciclo agricola)          │       │
│  │ Monitoreo → Riesgo → Estado →          │       │
│  │ Accion → Priorizacion → Registro       │       │
│  └────────────────────────────────────────┘       │
├──────────────────────────────────────────────────┤
│  SDK decision-engine/core/                        │
│  (decisions + explain + validate + audit)         │
├──────────────────────────────────────────────────┤
│  yarumo compute/ + maths/                       │
│  (deductive, bayesian, fuzzy, causal, mcdm)      │
├──────────────────────────────────────────────────┤
│  Integraciones externas                           │
│  ┌────────────┬───────────┬───────────────────┐  │
│  │ IoT        │ Clima     │ Conocimiento      │  │
│  │(sensores,  │(IDEAM,    │(Cenicafe, ICA,    │  │
│  │ estaciones)│ pronost.) │ gremios)          │  │
│  └────────────┴───────────┴───────────────────┘  │
├──────────────────────────────────────────────────┤
│  Storage + Telemetry                              │
└──────────────────────────────────────────────────┘
```

---

## 11) Go-to-market

**Fase 1**: Piloto con FNC (Federacion de Cafeteros) en una zona cafetera. ~50-100 fincas en municipio con alta incidencia de roya. Implementar: monitoreo IoT basico (estacion climatica por vereda) + bayesian roya + deductive manejo + app movil con recomendaciones. Demostrar deteccion anticipada de roya vs alertas genericas de Cenicafe.

**Fase 2**: Escalar a 500-1,000 fincas. Agregar fuzzy estado de cultivo + MCDM para plan de renovacion. Medir: reduccion de perdidas por roya, adopcion de recomendaciones, registro automatico de BPA.

**Fase 3**: Saltar a flores (Asocolflores) — invernaderos con IoT denso. Diferente cultivo, mismo motor. Control ambiental + calidad de corte + trazabilidad de exportacion.

**Fase 4**: Banano — emergencia Fusarium R4T. Pitch al ICA/gobierno: "deteccion temprana bayesian + protocolo de cuarentena automatizado". Financiamiento publico/cooperacion internacional (FAO, USAID ya financian programas Fusarium en LATAM).

**Fase 5**: Plataforma multi-cultivo para agroexportadores. La certificacion (GlobalGAP, Rainforest Alliance) es el gancho — AgroDecide genera el registro automaticamente.

---

## 12) Por que 75% y no 100%

| Gap | Que es | Como se resuelve | Impacto |
|---|---|---|---|
| **Infraestructura IoT** | Sensores, conectividad en zona rural, energia | Hardware + telecom, completamente fuera de scope | **Alto** |
| **Datos meteorologicos** | Pronostico localizado, datos IDEAM | Integracion API, no inferencia | **Medio** |
| **Imagenes satelitales/drones** | NDVI, deteccion visual de estres | CV/ML externo, fuera de scope | **Medio** |
| **Conocimiento agronomico formalizado** | Reglas de manejo por cultivo, umbrales | Trabajo de dominio (agronomos + ingenieros), como ClinicalRules con GPC | **Alto** |
| **Conectividad rural** | Muchas zonas cafeteras sin 4G | App offline-first con sync, infra no inferencia | **Medio** |
| **Adopcion del agricultor** | El cafetero de 60 anos no usa app | UX extremadamente simple, SMS como fallback, extensionista como intermediario | **Medio** |

Los 2 gaps altos (IoT + conocimiento agronomico) son similares a ClinicalRules (FHIR + base de conocimiento clinico): no son limitaciones del motor, son barreras de entrada al dominio.

---

## 13) Por que ★★★☆☆ y no mas

| Factor | ComplianceEngine (★★★★★) | AgroDecide (★★★☆☆) |
|---|---|---|
| Barrera de entrada | Baja — reglas son texto legal | Alta — necesita agronomos + IoT + clima |
| Complejidad de datos | Baja — datos estructurados | Alta — sensores, clima, imagenes, heterogeneo |
| Infraestructura requerida | Existente (internet, computadores) | **Debe crearse** (sensores, conectividad rural) |
| Regulacion de producto | No aplica | Baja (no es dispositivo regulado) |
| Time-to-market | 3-6 meses | 6-12 meses (IoT + conocimiento + piloto campo) |
| Capacidad de pago del cliente | Alta (entidades vigiladas) | **Baja** (pequeno agricultor) — depende de gremios/gobierno |
| Impacto social | Medio | **Alto** — 540K familias cafeteras, seguridad alimentaria |

La paradoja: AgroDecide tiene el impacto social mas alto de todos los productos pero la capacidad de pago mas baja del cliente final. La solucion es B2B2C: vender al gremio/gobierno, servir al agricultor.
