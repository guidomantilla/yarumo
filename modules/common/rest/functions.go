package rest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

const (
	applicationJSON           = "application/json"
	applicationFormURLEncoded = "application/x-www-form-urlencoded"
)

// Call executes a REST request described by spec and returns the decoded response.
func Call[T any](ctx context.Context, spec *RequestSpec, options ...Option) (*ResponseSpec[T], error) {
	if cutils.Nil(ctx) {
		return nil, ErrCall(ErrContextNil)
	}

	if cutils.Nil(spec) {
		return nil, ErrCall(ErrRequestSpecNil)
	}

	opts := NewOptions(options...)

	req, err := spec.Build(ctx)
	if err != nil {
		return nil, ErrCall(err)
	}

	start := time.Now()

	resp, err := opts.doFn(req)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	if err != nil {
		return nil, ErrCall(err)
	}

	body, err := readResponseBody(resp.Body, opts.maxResponseSize)
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

// CallStream executes a REST request and returns the raw response body for streaming consumption.
// The caller is responsible for closing Body on the returned StreamResponseSpec.
func CallStream(ctx context.Context, spec *RequestSpec, options ...Option) (*StreamResponseSpec, error) {
	if cutils.Nil(ctx) {
		return nil, ErrCall(ErrContextNil)
	}

	if cutils.Nil(spec) {
		return nil, ErrCall(ErrRequestSpecNil)
	}

	opts := NewOptions(options...)

	req, err := spec.Build(ctx)
	if err != nil {
		return nil, ErrCall(err)
	}

	resp, err := opts.doFn(req)
	if err != nil {
		if resp != nil {
			_ = resp.Body.Close()
		}

		return nil, ErrCall(err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, readErr := readResponseBody(resp.Body, opts.maxResponseSize)
		_ = resp.Body.Close()

		if readErr != nil {
			return nil, ErrCall(readErr)
		}

		return nil, ErrCall(&HTTPError{StatusCode: resp.StatusCode, Status: resp.Status, Body: body})
	}

	return &StreamResponseSpec{
		Headers: resp.Header.Clone(),
		Code:    resp.StatusCode,
		Status:  resp.Status,
		Body:    resp.Body,
	}, nil
}

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

// DecodeHTTPError extracts and decodes the JSON body of an HTTPError into the target type E.
func DecodeHTTPError[E any](err error) (E, bool) {
	var zero E

	var httpErr *HTTPError

	ok := errors.As(err, &httpErr)
	if !ok {
		return zero, false
	}

	if cutils.Empty(httpErr.Body) {
		return zero, false
	}

	var decoded E

	unmarshalErr := json.Unmarshal(httpErr.Body, &decoded)
	if unmarshalErr != nil {
		return zero, false
	}

	return decoded, true
}
