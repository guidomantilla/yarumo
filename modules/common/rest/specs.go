package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/guidomantilla/yarumo/common/assert"
)

type ResponseSpec[T any] struct {
	Duration      time.Duration
	ContentLength int64
	Headers       map[string][]string
	Code          int
	Status        string
	Body          T
}

type RequestSpec struct {
	Method      string
	URL         string
	Path        string
	Headers     map[string]string
	QueryParams map[string][]string
	// Deprecated: use Body instead
	RawBody []byte
	Body    any
}

func (spec *RequestSpec) Build(ctx context.Context) (*http.Request, error) {
	assert.NotEmpty(spec, "request spec is nil")
	assert.NotEmpty(ctx, "ctx is nil")

	// Fix headers

	if spec.Headers == nil {
		spec.Headers = make(map[string]string)
	}
	_, ok := spec.Headers["Accept"]
	if !ok {
		spec.Headers["Accept"] = "application/json"
	}

	// Process body

	var err error
	var body io.Reader
	if spec.Body != nil {
		spec.RawBody, err = json.Marshal(spec.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(spec.RawBody)

		spec.Headers["Content-Length"] = fmt.Sprintf("%d", len(spec.RawBody))
		_, ok := spec.Headers["Content-Type"]
		if !ok {
			spec.Headers["Content-Type"] = "application/json"
		}
	}

	// Build request

	u, err := url.Parse(spec.URL)
	if err != nil {
		return nil, err
	}

	if spec.Path != "" {
		u.Path = path.Join(u.Path, spec.Path)
	}

	q := u.Query()
	for k, vals := range spec.QueryParams {
		for _, v := range vals {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, spec.Method, u.String(), body)
	if err != nil {
		return nil, err
	}

	for k, v := range spec.Headers {
		req.Header.Set(k, v)
	}

	return req, nil
}
