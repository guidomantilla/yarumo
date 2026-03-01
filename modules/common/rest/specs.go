package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
	ctypes "github.com/guidomantilla/yarumo/common/types"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

// ResponseSpec holds the decoded response from a REST call.
type ResponseSpec[T any] struct {
	Duration      time.Duration
	ContentLength int64
	Headers       map[string][]string
	Code          int
	Status        string
	Body          T
}

// StreamResponseSpec holds the raw streaming response from a REST call.
// The caller is responsible for closing Body.
type StreamResponseSpec struct {
	Headers map[string][]string
	Code    int
	Status  string
	Body    io.ReadCloser
}

// RequestSpec describes a REST request to be built and executed.
type RequestSpec struct {
	Method      string
	URL         string
	Path        string
	Headers     map[string]string
	QueryParams map[string][]string
	// Deprecated: use Body instead.
	RawBody ctypes.Bytes
	Body    any
}

// Build constructs an *http.Request from the spec without mutating the receiver.
func (spec *RequestSpec) Build(ctx context.Context) (*http.Request, error) {
	cassert.NotNil(spec, "request spec is nil")

	if cutils.Nil(ctx) {
		return nil, ErrCall(ErrContextNil)
	}

	bodyReader, bodyLen, contentType, err := spec.marshalBody()
	if err != nil {
		return nil, ErrCall(err)
	}

	headers := spec.buildHeaders(bodyReader != nil, contentType)
	if bodyLen > 0 {
		headers["Content-Length"] = strconv.Itoa(bodyLen)
	}

	reqURL, err := spec.buildURL()
	if err != nil {
		return nil, ErrCall(err)
	}

	req, err := http.NewRequestWithContext(ctx, spec.Method, reqURL, bodyReader)
	if err != nil {
		return nil, ErrCall(err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

// buildHeaders returns a copy of the spec headers with defaults applied.
func (spec *RequestSpec) buildHeaders(hasBody bool, bodyContentType string) map[string]string {
	headers := make(map[string]string, len(spec.Headers)+2)
	maps.Copy(headers, spec.Headers)

	_, hasAccept := headers["Accept"]
	if !hasAccept {
		headers["Accept"] = applicationJSON
	}

	if hasBody && bodyContentType != "" {
		_, hasCT := headers["Content-Type"]
		if !hasCT {
			headers["Content-Type"] = bodyContentType
		}
	}

	return headers
}

// marshalBody returns a reader, size, inferred content-type, and error for the request body.
func (spec *RequestSpec) marshalBody() (io.Reader, int, string, error) {
	if spec.Body == nil {
		return nil, 0, "", nil
	}

	switch v := spec.Body.(type) {
	case io.Reader:
		return v, 0, "", nil
	case []byte:
		return bytes.NewReader(v), len(v), "", nil
	case string:
		return strings.NewReader(v), len(v), "", nil
	case url.Values:
		encoded := v.Encode()

		return strings.NewReader(encoded), len(encoded), applicationFormURLEncoded, nil
	default:
		raw, err := json.Marshal(spec.Body)
		if err != nil {
			return nil, 0, "", cerrs.Wrap(ErrMarshalBodyFailed, err)
		}

		return bytes.NewReader(raw), len(raw), applicationJSON, nil
	}
}

// buildURL combines the spec URL, path, and query parameters into a final URL string.
func (spec *RequestSpec) buildURL() (string, error) {
	u, err := url.Parse(spec.URL)
	if err != nil {
		return "", cerrs.Wrap(ErrURLParseFailed, err)
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

	return u.String(), nil
}
