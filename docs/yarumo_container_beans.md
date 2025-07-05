# Beans recomendados para un contenedor tipo Spring Boot en Go

Tu contenedor actual en Go busca replicar un `ApplicationContext` como el de Spring Boot. Aquí tienes una guía organizada por categorías con beans adicionales sugeridos.

---

## ✅ 1. Infraestructura básica

| Campo sugerido  | Tipo / Paquete                      | Propósito                                         |
|------------------|-------------------------------------|---------------------------------------------------|
| `Tracer`         | `otel.Tracer`                      | Observabilidad y tracing distribuido              |
| `Metrics`        | `prometheus.Registry` / `otel.Meter` | Métricas e instrumentación                        |
| `EventBus`       | `yarumo/event.Bus` o similar        | Comunicación pub/sub interna                      |
| `Clock`          | `yarumo/time.Clock` (interfaz)      | Inyección de tiempo controlado para pruebas       |
| `IDGenerator`    | `func() string`                     | Generador de IDs (UUID, ULID, etc.)               |

---

## ✅ 2. Acceso a datos / persistencia

| Campo sugerido       | Tipo                          | Propósito                                     |
|----------------------|-------------------------------|-----------------------------------------------|
| `DB`                 | `*sql.DB`, `gorm.DB`, etc.    | Acceso a base de datos relacional             |
| `Redis`              | `*redis.Client`               | Caché, rate limit, sesiones                   |
| `TransactionManager` | Interfaz propia               | Manejo explícito de transacciones             |

---

## ✅ 3. Configuración dinámica / runtime

| Campo sugerido   | Tipo                           | Propósito                                         |
|------------------|--------------------------------|---------------------------------------------------|
| `Env`            | `map[string]string` / struct   | Variables de entorno o configuración general     |
| `FeatureFlags`   | `yarumo/feature.ToggleManager` | Control dinámico de funcionalidades (feature flags) |

---

## ✅ 4. Comunicación externa

| Campo sugerido    | Tipo                              | Propósito                                         |
|-------------------|-----------------------------------|---------------------------------------------------|
| `HTTPClient`      | `*http.Client`                    | Cliente HTTP con configuraciones reutilizables   |
| `MessageBroker`   | Interface para Kafka/NATS/etc.    | Comunicación asíncrona entre servicios           |
| `EmailSender`     | Wrapper para SMTP/API             | Notificaciones por correo electrónico            |

---

## ✅ 5. Seguridad / Identidad

| Campo sugerido        | Tipo                      | Propósito                                      |
|------------------------|---------------------------|------------------------------------------------|
| `TokenVerifier`        | `tokens.Verifier`         | Verificación de tokens (JWT, etc.)             |
| `AuthContextProvider`  | Interfaz de usuario actual| Abstracción del usuario autenticado            |

---

## ✅ 6. Aplicación

| Campo sugerido      | Tipo                           | Propósito                                         |
|---------------------|--------------------------------|---------------------------------------------------|
| `Router`            | `chi.Router`, `mux.Router`     | Ruteo HTTP                                       |
| `LifecycleManager`  | `yarumo/lifecycle.Manager`     | Hooks `OnStart`, `OnStop` y shutdown ordenado    |

---

## 🧩 Ejemplo de Container expandido

```go
type Container struct {
  AppName     string
  AppVersion  string
  Config      any
  Logger      zerolog.Logger
  Validator   *validator.Validate

  // Seguridad
  PasswordEncoder   passwords.Encoder
  PasswordGenerator passwords.Generator
  TokenGenerator    tokens.Generator
  TokenVerifier     tokens.Verifier

  // Observabilidad
  Tracer     trace.Tracer
  Metrics    prometheus.Registerer

  // Infraestructura
  DB         *sql.DB
  Redis      *redis.Client
  EventBus   event.Bus
  Clock      yarumo.Clock
  IDGen      func() string

  // Comunicación
  HTTPClient     *http.Client
  MessageBroker  pubsub.Broker

  // Config/Env
  Env          map[string]string
  FeatureFlags yarumo.FeatureFlagManager
}
```
