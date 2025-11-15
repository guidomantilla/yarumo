package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

type ResponseSpec[T any] struct {
	Duration      time.Duration
	ContentLength int64
	Headers       map[string][]string
	Code          int
	Status        string
	RawBody       []byte
	Body          T
}

type RequestSpec struct {
	Method      string
	URL         string
	Path        string
	Headers     map[string]string
	QueryParams map[string]string
	RawBody     []byte
	Body        any
}

func (spec *RequestSpec) Build(ctx context.Context) (*http.Request, error) {

	u, err := url.Parse(spec.URL)
	if err != nil {
		return nil, err
	}

	if spec.Path != "" {
		u.Path = path.Join(u.Path, spec.Path)
	}

	q := u.Query()
	for k, v := range spec.QueryParams {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	var body io.Reader
	if spec.Body != nil {
		spec.RawBody, err = json.Marshal(spec.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(spec.RawBody)
	}

	req, err := http.NewRequestWithContext(ctx, spec.Method, u.String(), body)
	if err != nil {
		return nil, err
	}

	for k, v := range spec.Headers {
		req.Header.Set(k, v)
	}

	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}

	return req, nil
}
