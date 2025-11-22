package rest

import (
	"encoding/json"
	"mime"
	"net/http"
	"strings"

	"github.com/guidomantilla/yarumo/common/utils"
)

// isJSONMediaType checks if the media type is JSON.
func isJSONMediaType(mediaType string) bool {
	if utils.Equal(mediaType, "application/json") {
		return true
	}
	return strings.HasSuffix(mediaType, "+json")
}

// decodeResponseBody decodes the response body into the provided type.
func decodeResponseBody[T any](body []byte, statusCode int, contentType string) (T, error) {
	var zero T
	if utils.Empty(http.StatusText(statusCode)) || utils.Equal(http.StatusNoContent, statusCode) || utils.Empty(body) {
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
	if isJSONMediaType(mediaType) || mediaType == "" {
		err := json.Unmarshal(body, &decoded)
		if err != nil {
			return zero, err
		}
		return decoded, nil
	}

	return zero, &DecodeResponseError[T]{ContentType: contentType, T: zero}
}
