package rest

import (
	"encoding/json"
	"fmt"
	"mime"
	"strings"
)

// IsJSONMediaType informa si el media type es JSON (application/json o *+json)
func IsJSONMediaType(mediaType string) bool {
	if mediaType == "application/json" {
		return true
	}
	return strings.HasSuffix(mediaType, "+json")
}

// DecodeResponseBody extrae la lógica de decodificación genérica del cuerpo HTTP.
func DecodeResponseBody[T any](body []byte, statusCode int, contentType string) (T, error) {
	var zero T
	if statusCode == 204 || len(body) == 0 {
		return zero, nil
	}

	mediaType := strings.TrimSpace(contentType)
	if mediaType != "" {
		mt, _, err := mime.ParseMediaType(contentType)
		if err == nil {
			mediaType = mt
		}
	}

	_, ok := any(*new(T)).([]byte)
	if ok {
		return any(body).(T), nil
	}

	_, ok = any(*new(T)).(string)
	if ok {
		return any(string(body)).(T), nil
	}

	var decoded T
	if IsJSONMediaType(mediaType) || mediaType == "" {
		err := json.Unmarshal(body, &decoded)
		if err != nil {
			return zero, err
		}
		return decoded, nil
	}

	return zero, fmt.Errorf("content-type '%s' no soportado para el tipo de respuesta solicitado", contentType)
}
