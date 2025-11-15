### Revisión nuevamente (con hallazgos adicionales)
He releído tus archivos `rest/*` y también el cliente HTTP base en `modules/common/http/client.go`. Además de los puntos que ya te comenté, hay un par de detalles importantes que surgen al ver cómo implementaste `retries` y `rate limiting` en el cliente HTTP y cómo eso se integra con tu capa `rest`.

A continuación dejo la lista consolidada, marcando lo nuevo/crítico.

---

### [x] 1) Retries con cuerpo de request: riesgo de enviar cuerpo vacío en reintentos (crítico)
- Dónde: `modules/common/http/client.go`, método `(*client).Do` (líneas 52–73).
- Problema: reintentas la misma `*http.Request` en un bucle de `retry.DoWithData`. Si la request tiene body, después del primer intento `req.Body` queda consumido. El `http.Transport` puede usar `req.GetBody` para reintentos automáticos propios, pero en tu bucle manual de `retry` no estás reseteando el body antes de cada nuevo `Do`. Resultado: en reintentos manuales podrías mandar un cuerpo vacío.
- Por qué te afecta menos en tu caso: en `rest/specs.go` usas `bytes.NewReader(...)` para crear el body. `http.NewRequestWithContext` detecta lectores re-seekables y puede establecer `GetBody` internamente. Aun así, para los reintentos manuales del cliente sería prudente restablecer explícitamente `req.Body` usando `req.GetBody` antes de cada `Do` cuando `req.Body` exista. Esto te protege ante futuros cambios o flujos distintos.
- Sugerencia concreta dentro del closure `retryableCall`:
    - Antes de `c.Client.Do(req)`, si `req.GetBody != nil`, hacer `rc, _ := req.GetBody(); req.Body = rc`.
    - Si `req.GetBody == nil` y `req.Body != nil`, documenta que no se pueden reintentar requests con body no re-creable.

### [x] 2) Retries por códigos HTTP (5xx, 429) no ocurren (diseño)
- Dónde: `modules/common/http/client.go` `Do` (líneas 54–73).
- Problema: tu retry actual sólo reintenta ante errores de transporte (cuando `err != nil`). Si el servidor devuelve `HTTP 5xx` o `429`, `err` es `nil` y tu `retryIf` no se entera, por lo que no hay reintentos en base a códigos de estado.
- Opciones:
    - a) Mover la lógica de decidir reintentos por status code a este cliente HTTP: tras `c.Client.Do(req)`, inspeccionar `res.StatusCode` y si coincide con política de retry, cerrar `res.Body` y devolver un error sentinela (envolviendo el código) para que `retry` lo procese. Tu `retryIf` puede comprobar ese sentinela.
    - b) Extender las `Options` para permitir un `RetryOnResponse func(*http.Response) (bool, error)` y en caso de `true`, devolver un error para disparar el retry.
    - c) Alternativamente, mover esa política a la capa `rest.Call` (menos ideal si quieres que el cliente HTTP sea reutilizable).

### [x] 3) Duración medida antes de leer el cuerpo (precisión)
- Dónde: `modules/common/rest/client.go` (líneas 19–33).
- Problema: estás midiendo `duration` justo tras `Do(req)`, antes de `io.ReadAll`. Así mides TTFB y headers, no el tiempo total.
- Corrección: calcula `duration` después de leer el body (y opcionalmente después de decodificar JSON).

### [x] 4) `ResponseSpec.RawBody` no se rellena (usabilidad)
- Dónde: `modules/common/rest/client.go` retorno.
- Problema: `specs.go` define `ResponseSpec.RawBody`, pero nunca lo llenas.
- Corrección: asigna `RawBody: body` en la respuesta. Esto ayuda a depuración y a casos en que no quieres decodificar.

### [x] 5) `ContentLength` real vs `resp.ContentLength` (detalle)
- Dónde: `modules/common/rest/client.go` retorno.
- Comentario: `resp.ContentLength` puede ser `-1`. Como ya tienes el `body`, puedes usar `int64(len(body))` cuando el `ContentLength` "de red" sea desconocido.

### [x] 6) Decodificación incondicional a JSON (flexibilidad)
- Dónde: `modules/common/rest/client.go` (líneas 38–44).
- Problema: siempre intentas `json.Unmarshal` si hay body. Si la respuesta no es JSON, fallarás.
- Mejora: usa `Content-Type` para decidir. Y soporta `T` comunes como `[]byte` o `string` sin `Unmarshal`.

### [x] 7) `Build` muta `RequestSpec` (sorpresa de API)
- Dónde: `modules/common/rest/specs.go` (líneas 51–58).
- Problema: `Build` muta `spec.RawBody` cuando serializa `Body`. Considera no mutar, o documentarlo claramente.

### [] 8) Prioridad entre `RawBody` y `Body` (robustez)
- Dónde: `modules/common/rest/specs.go`.
- Mejora: si `RawBody` viene dado, úsalo tal cual. Sólo serializa `Body` si `RawBody` está vacío.

### [x] 9) `Content-Type` por defecto sólo cuando hay body
- Dónde: `modules/common/rest/specs.go` (líneas 69–74).
- Mejora: no fuerces `Content-Type` en GET sin body; deja sólo `Accept: application/json` por defecto.

### [] 10) Query params repetibles
- Dónde: `modules/common/rest/specs.go` (líneas 45–49).
- Mejora: cambia `map[string]string` por `map[string][]string` para soportar `a=1&a=2`.

### [x] 11) `Error.Unwrap` para compatibilidad con `errors.Is/As`
- Dónde: `modules/common/rest/errors.go`.
- Mejora: añade `Unwrap() error { return e.Err }` a tu `Error`. Así podrás hacer `errors.As(err, &rest.HTTPError)` cómodamente.

### [] 12) Extras
- `path.Join` puede normalizar barras; valida que sea el comportamiento deseado.
- Considera una variante `CallStream` para respuestas grandes.
- Si planeas `retry/backoff` por status en la capa HTTP, asegúrate de cerrar `res.Body` antes de cada reintento para evitar fugas.

---

### Sugerencia de ajustes mínimos
- En `rest/client.go`:
    - [x] Mover el cálculo de `duration` tras `io.ReadAll`.
    - [] Rellenar `RawBody`.
    - [] Usar `len(body)` cuando `resp.ContentLength < 0`.
    - [] Opcional: decodificar basado en `Content-Type` o soportar `[]byte`/`string`.

- En `rest/specs.go`:
    - [] Dar prioridad a `RawBody` si viene.
    - [] No mutar `spec` o documentarlo.
    - [] No forzar `Content-Type` sin body; mantener `Accept`.
    - [] Considerar `QueryParams map[string][]string`.

- En `http/client.go`:
    - [] Antes de cada `Do`, si `req.GetBody != nil`, re-crear `req.Body` con `req.GetBody()` para que los reintentos envíen el mismo cuerpo.
    - [] Añadir política opcional de retry por status (`5xx`, `429`) devolviendo un error sentinela si aplica, y cerrando previamente `res.Body`.

Si quieres, preparo un patch mínimo aplicando estos cambios clave manteniendo tu API y compatibilidad con tests actuales.