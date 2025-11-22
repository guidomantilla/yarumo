package rest

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/utils"
)

// Call executes a request and returns the response.
func Call[T any](ctx context.Context, spec *RequestSpec, options ...Option) (*ResponseSpec[T], error) {
	assert.NotEmpty(ctx, "ctx is nil")
	assert.NotEmpty(spec, "spec is nil")

	opts := NewOptions(options...)
	req, err := spec.Build(ctx)
	if err != nil {
		return nil, ErrCall(err)
	}

	start := time.Now()
	resp, err := opts.DoFn(req)
	if err != nil {
		return nil, ErrCall(err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrCall(err)
	}

	duration := time.Since(start)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ErrCall(&HTTPError{StatusCode: resp.StatusCode, Status: resp.Status, Body: body})
	}

	decoded, err := decodeResponseBody[T](body, resp.StatusCode, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, ErrCall(err)
	}

	return &ResponseSpec[T]{
		Duration:      duration,
		ContentLength: int64(len(body)),
		Headers:       resp.Header.Clone(),
		Code:          resp.StatusCode,
		Status:        resp.Status,
		Body:          decoded,
	}, nil
}

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
