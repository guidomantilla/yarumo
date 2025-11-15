package rest

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/guidomantilla/yarumo/common/http"
)

func Call[T any](ctx context.Context, client http.Client, spec RequestSpec) (*ResponseSpec[T], error) {

	req, err := spec.Build(ctx)
	if err != nil {
		return nil, ErrCall(err)
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrCall(err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	duration := time.Since(start)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrCall(err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ErrCall(&HTTPError{StatusCode: resp.StatusCode, Status: resp.Status, Body: body})
	}

	var decoded T
	if len(body) > 0 {
		err = json.Unmarshal(body, &decoded)
		if err != nil {
			return nil, ErrCall(err)
		}
	}

	return &ResponseSpec[T]{
		Duration:      duration,
		ContentLength: resp.ContentLength,
		Headers:       resp.Header.Clone(),
		Code:          resp.StatusCode,
		Status:        resp.Status,
		Body:          decoded,
	}, nil
}
