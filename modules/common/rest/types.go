package rest

import "context"

var (
	_ CallFn[any] = Call
)

type CallFn[T any] func(ctx context.Context, spec *RequestSpec, options ...Option) (*ResponseSpec[T], error)
