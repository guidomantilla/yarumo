package comm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/guidomantilla/yarumo/pkg/common/log"
	"github.com/guidomantilla/yarumo/pkg/common/pointer"
)

// Http

func AppendBody(h http.Header, key string, body []byte) log.EventFn {
	return func(e *zerolog.Event) {
		if JsonPayload(h) {
			e.RawJSON(key, body)
		} else {
			e.Bytes(key, body)
		}
	}
}

func JsonPayload(h http.Header) bool {
	contentType := h.Get("Content-Type")
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	return strings.HasPrefix(contentType, "application/json") || strings.Contains(contentType, "+json")
}

func MustJsonMarshalSanitized(h http.Header) []byte {
	b, err := JsonMarshalSanitized(h)
	if err != nil {
		return []byte(fmt.Sprintf("error marshalling header: %s", err.Error()))
	}
	return b
}

func JsonMarshalSanitized(h http.Header) ([]byte, error) {
	sanitized := make(map[string][]string, len(h))

	for k, values := range h {
		safeKey := strings.ToValidUTF8(k, "")
		safeValues := make([]string, len(values))

		for i, v := range values {
			safeValues[i] = strings.ToValidUTF8(v, "")
		}

		sanitized[safeKey] = safeValues
	}

	return json.Marshal(sanitized)
}

func ToReadNopCloser(reader io.ReadCloser) (io.ReadCloser, []byte, error) {
	if reader == nil {
		return nil, nil, fmt.Errorf("nil reader")
	}

	buffer, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading from reader: %w", err)
	}

	readerNopCloser := io.NopCloser(bytes.NewReader(buffer))
	return readerNopCloser, buffer, nil
}

// Rest

func ToReader(body []byte) io.Reader {
	if body == nil {
		return nil
	}
	return io.NopCloser(bytes.NewReader(body))
}

func MarshalRequest(body any) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %w", err)
	}

	return ToReader(data), nil
}

func UnmarshalResponse[T any](body io.Reader) (T, error) {
	if body == nil {
		return pointer.Zero[T](), nil
	}

	var response T
	err := json.NewDecoder(body).Decode(&response)
	if err != nil && err != io.EOF {
		return pointer.Zero[T](), fmt.Errorf("error unmarshalling response body: %w", err)
	}
	return response, nil
}

//

func ToSliceOfMapsOfAny(input any) ([]map[string]any, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	slice, ok := input.([]any)
	if !ok {
		return nil, fmt.Errorf("input is not a slice of any, got %T", input)
	}

	result := make([]map[string]any, 0, len(slice))
	for i, item := range slice {
		m, err := ToMapOfAny(item)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("element at index %d is not a map[string]any", i))
		}
		result = append(result, m)
	}
	return result, nil
}

func ToMapOfAny(input any) (map[string]any, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	m, ok := input.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("input is not a map[string]any, got %T", input)
	}
	return m, nil
}
