package tokens

import (
	"github.com/stretchr/testify/mock"
)

var (
	_ Generator = (*MockGenerator)(nil)
)

type MockGenerator struct {
	mock.Mock
}

func NewMockGenerator() *MockGenerator {
	return &MockGenerator{}
}

func (mock *MockGenerator) Generate(subject string, principal Principal) (*string, error) {
	args := mock.Called(subject, principal)
	return args.Get(0).(*string), args.Error(1)
}

func (mock *MockGenerator) Validate(tokenString string) (Principal, error) {
	args := mock.Called(tokenString)
	return args.Get(0).(Principal), args.Error(1)
}
