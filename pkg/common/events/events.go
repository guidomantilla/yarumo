package events

import (
	"context"
	"sync"

	"github.com/guidomantilla/yarumo/pkg/common/utils"
)

type eventBus[T any] struct {
	subscribers map[EventType]map[string]Subscriber[T]
	mutex       sync.Mutex
}

func NewEventBus[T any]() EventBus[T] {
	return &eventBus[T]{
		subscribers: make(map[EventType]map[string]Subscriber[T]),
	}
}

func (eb *eventBus[T]) Subscribe(_ context.Context, eventType EventType, subscriber Subscriber[T]) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if utils.Nil(eb.subscribers[eventType]) {
		eb.subscribers[eventType] = make(map[string]Subscriber[T])
	}

	eb.subscribers[eventType][subscriber.String()] = subscriber
}

func (eb *eventBus[T]) Unsubscribe(_ context.Context, eventType EventType, subscriber Subscriber[T]) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if subscribers, ok := eb.subscribers[eventType]; ok {
		delete(subscribers, subscriber.String())

		if utils.Empty(subscribers) {
			delete(eb.subscribers, eventType)
		}
	}
}

func (eb *eventBus[T]) Publish(ctx context.Context, eventType EventType, event T) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if subscribers, ok := eb.subscribers[eventType]; ok {
		for _, subscriber := range subscribers {
			subscriber(ctx, event)
		}
	}
}
