package http

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (mock *MockClient) Do(req *http.Request, options ...Option) (*http.Response, error) {
	args := mock.Called(req, options)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (mock *MockClient) LimiterEnabled() bool {
	args := mock.Called()
	return args.Bool(0)
}
