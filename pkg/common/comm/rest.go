package comm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/guidomantilla/yarumo/pkg/common/pointer"
)

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

type RESTResponse[T any] struct {
	Code   int            `json:"code,omitempty"`
	Status string         `json:"status,omitempty"`
	Data   T              `json:"data,omitempty"`
	Error  map[string]any `json:"error,omitempty"`
}

func RESTCall[T any](ctx context.Context, method string, url string, body any, headers http.Header, opts ...RestOption) (*RESTResponse[T], error) {
	options := NewRestOptions(opts...)

	reader, err := MarshalRequest(body)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("error creating request object: %w", err)
	}

	req.Header = headers.Clone()
	req.Header.Set("Content-Type", "application/json")

	resp, err := options.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling %s %s: %w", method, url, err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	response := &RESTResponse[T]{
		Code:   resp.StatusCode,
		Status: http.StatusText(resp.StatusCode),
	}

	if resp.StatusCode >= 400 {
		data, err := UnmarshalResponse[map[string]any](resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling error response body: %w", err)
		}
		response.Error = data
		return response, nil
	}

	data, err := UnmarshalResponse[T](resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}
	response.Data = data
	return response, nil
}
