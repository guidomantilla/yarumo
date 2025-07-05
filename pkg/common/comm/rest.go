package comm

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/guidomantilla/yarumo/pkg/common/utils"
)

type restClient struct {
	url              string
	http             HTTPClient
	headers          http.Header
	statusCodeErrors []int
	statusCodeOK     []int
}

func NewRESTClient(url string, opts ...RestOption) RESTClient {
	options := NewRestOptions(opts...)
	return &restClient{
		url:              url,
		http:             options.http,
		headers:          options.headers,
		statusCodeErrors: options.statusCodeErrors,
		statusCodeOK:     options.statusCodeOK,
	}
}

func (rest *restClient) Call(ctx context.Context, method string, path string, body any) (*RESTResponse, error) {

	url := fmt.Sprintf("%s%s", rest.url, path)

	reader, err := MarshalRequest(body)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("error creating request object: %w", err)
	}

	req.Header = rest.headers.Clone()

	resp, err := rest.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling %s %s: %w", method, url, err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	response := &RESTResponse{
		Code:   resp.StatusCode,
		Status: http.StatusText(resp.StatusCode),
	}

	if utils.NotIn(resp.StatusCode, rest.statusCodeOK...) {
		data, err := UnmarshalResponse[map[string]any](resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling error response body: %w", err)
		}
		response.Error = data
		return response, nil
	}

	if utils.In(resp.StatusCode, rest.statusCodeOK...) {

		data, err := UnmarshalResponse[any](resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling response body: %w", err)
		}
		response.Data = data
		return response, nil
	}

	return nil, fmt.Errorf("cannot handle response with status code %d", resp.StatusCode)
}

//

type RESTResponse struct {
	Code   int            `json:"code,omitempty"`
	Status string         `json:"status,omitempty"`
	Data   any            `json:"data,omitempty"`
	Error  map[string]any `json:"error,omitempty"`
}
