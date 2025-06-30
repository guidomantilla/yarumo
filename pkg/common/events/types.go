package events

import (
	"context"
	"fmt"
)

type EventType string

type Subscriber[T any] func(ctx context.Context, event T)

func (s Subscriber[T]) String() string {
	return fmt.Sprintf("%T", s)
}

//

var (
	_ EventBus[any] = (*eventBus[any])(nil)
)

type EventBus[T any] interface {
	Subscribe(ctx context.Context, eventType EventType, subscriber Subscriber[T])
	Unsubscribe(ctx context.Context, ventType EventType, subscriber Subscriber[T])
	Publish(ctx context.Context, eventType EventType, event T)
}
