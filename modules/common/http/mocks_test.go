package http

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

var (
	_ Client = (*MockClient)(nil)
)

type MockClient struct {
	mock.Mock
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (mock *MockClient) Do(req *http.Request) (*http.Response, error) {
	args := mock.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (mock *MockClient) LimiterEnabled() bool {
	args := mock.Called()
	return args.Bool(0)
}

func (mock *MockClient) RetrierEnabled() bool {
	args := mock.Called()
	return args.Bool(0)
}
