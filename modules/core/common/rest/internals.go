package rest

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
	cutils "github.com/guidomantilla/yarumo/core/common/utils"
)

// readResponseBody reads the response body up to maxSize bytes and returns an error if exceeded.
func readResponseBody(body io.Reader, maxSize int64) ([]byte, error) {
	limited := io.LimitReader(body, maxSize+1)

	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, cerrs.Wrap(ErrReadBodyFailed, err)
	}

	if int64(len(data)) > maxSize {
		return nil, ErrResponseTooLarge
	}

	return data, nil
}

// isJSONMediaType checks if the media type is JSON or a JSON variant.
func isJSONMediaType(mediaType string) bool {
	if cutils.Equal(mediaType, applicationJSON) {
		return true
	}

	return strings.HasSuffix(mediaType, "+json")
}

// parseMediaType extracts the media type from a content-type header value.
func parseMediaType(contentType string) string {
	mediaType := strings.TrimSpace(contentType)
	if mediaType == "" {
		return ""
	}

	mt, _, err := mime.ParseMediaType(contentType)
	if err == nil {
		mediaType = mt
	}

	return mediaType
}

// decodeResponseBody decodes the response body into the provided type.
func decodeResponseBody[T any](body []byte, statusCode int, contentType string) (T, error) {
	var zero T

	if cutils.Empty(http.StatusText(statusCode)) || cutils.Equal(http.StatusNoContent, statusCode) || cutils.Empty(body) {
		return zero, nil
	}

	val, ok := any(body).(T)
	if ok {
		return val, nil
	}

	strVal, ok := any(string(body)).(T)
	if ok {
		return strVal, nil
	}

	mediaType := parseMediaType(contentType)
	if isJSONMediaType(mediaType) || mediaType == "" {
		var decoded T

		err := json.Unmarshal(body, &decoded)
		if err != nil {
			return zero, cerrs.Wrap(ErrUnmarshalFailed, err)
		}

		return decoded, nil
	}

	return zero, &DecodeResponseError[T]{ContentType: contentType}
}
