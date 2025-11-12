# Plan de Evolución del Wrapper HTTP en Go

Este documento describe el plan por fases para fortalecer, ampliar y profesionalizar el wrapper HTTP basado en `http.Client` de la librería estándar de Go, que ya implementa un `Do` con `rate.Limiter`.

---

## Fase 0 — Alineación y Hardening

- Revisar contrato de interfaz (`Client`), garantizando compatibilidad y claridad.
- Validar semántica del `rate.Limiter` (`Inf` = deshabilitado, `burst` > 0).
- Confirmar correcto cierre de `Body` en todas las rutas (éxito/error/retry).
- Alinear timeouts (`Client.Timeout`, `Transport` y `Context`).
- Homogeneizar errores (`ErrDoCall`, wrapping, causas y categorías).
- Añadir tests básicos (éxito, timeout, cancelación, limiter, retry).

---

## Fase 1 — Resiliencia y políticas de reintento

- Definir política de retry por tipo de error (red, 5xx, 429, etc.).
- Soportar `Retry-After` y backoff exponencial con jitter.
- Permitir reintentos seguros (`Idempotency-Key`).
- Incorporar `circuit breaker` opcional.
- Pruebas de estrés, caos y latencia.

---

## Fase 2 — Observabilidad

- Logging estructurado (zerolog): requests/responses redactadas.
- Métricas (Prometheus u OpenTelemetry): latencia, códigos HTTP, retries, esperas de limiter.
- Trazas distribuidas: spans por request, contexto de trace propagado.
- Hooks configurables desde `Options`.
- Tests de validación de métricas y trazas.

---

## Fase 3 — Ergonomía y Extensibilidad

- Permitir overrides de opciones por request (rate, retry, headers, timeout).
- Middlewares encadenables (`RoundTripper` stack).
- Builders por perfil (externo, interno, descargas).
- Validación de `Options` y defaults seguros.
- Documentación de comportamiento por defecto y overrides.

---

## Fase 4 — Casos de uso específicos

- REST a terceros: rate por host/endpoint, respeto de cuotas y `Retry-After`.
- Autenticación: bearer/JWT/API key con refresco automático.
- Headers especializados (`X-Request-Id`, `ETag`, versionado).
- Descargas: soporte `Range`, límite de bytes/segundo.
- Subidas: multipart/presigned, checksums, control de conexiones.
- Tests e2e por escenario.

---

## Fase 5 — Orquestación y Concurrencia

- Worker pool con `rate.Limiter` compartido.
- Priorización (múltiples limiters).
- Backpressure y control de encolado.
- Reportes agregados de lotes.

---

## Fase 6 — Seguridad y Cumplimiento

- Redacción de secretos en logs.
- Configuración TLS segura (`Transport`).
- Sanitización de errores (sin PII/secretos).
- Tests de seguridad y cumplimiento.

---

## Fase 7 — Rendimiento y Estabilidad

- Ajuste de `Transport`: `MaxConnsPerHost`, `IdleConnTimeout`.
- Benchmarks bajo carga.
- Pruebas de contención y memoria (`pprof`).

---

## Fase 8 — Entregables y Mantenimiento

- Documentación técnica y guía de uso.
- Ejemplos de configuración y escenarios.
- Versionado semántico y CI con matriz de Go/OS.
- Contrato estable de interfaz (para mocks/tests).

---

## Matriz de Pruebas

- Éxito y errores controlados (4xx, 5xx, 429).
- Timeout y cancelación de contexto.
- Limiter activo/inactivo.
- Auth y refresco.
- Descargas grandes con `Range`.
- Observabilidad (logs, métricas, trazas).

---

## Criterios de Salida

- Tests y métricas integradas.
- Documentación completa.
- Defaults seguros y reproducibles.
- Benchmarks estables.

