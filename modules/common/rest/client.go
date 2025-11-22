package rest

import (
	"context"
	"io"
	"time"

	"github.com/guidomantilla/yarumo/common/assert"
)

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
