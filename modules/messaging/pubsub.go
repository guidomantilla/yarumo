package messaging

import (
	"context"
	"reflect"
	"sync"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// ChannelFactory is the function type that builds a fresh Channel[T]
// on demand. The Pub/Sub facade calls it lazily the first time it
// needs a channel for a given Go type. The factory receives the
// reflect.Type for diagnostics (logging, name composition); it returns
// a Channel[any] because the facade stores channels in a heterogeneous
// registry. Implementations typically build a typed channel (e.g.
// NewPipelineChannel[Event]()) and wrap it in a type-erasing shim — the
// default factory does exactly that with PipelineChannel.
type ChannelFactory func(t reflect.Type) Channel[any]

// PubSub combines Publisher and Subscriber for in-process pub/sub
// routed by Go type. The same struct exposes Publish (via free
// function) and Subscribe (via free function) — see Publish[T] and
// Subscribe[T] below.
//
// PubSub maintains one Channel[any] per published Go type. The
// channels are allocated lazily on the first Publish or Subscribe for
// that type, using the configured ChannelFactory (default:
// NewPipelineChannel-backed). PubSub is safe for concurrent use.
type PubSub struct {
	factory ChannelFactory

	mu       sync.Mutex
	channels map[reflect.Type]Channel[any]
}

// Publisher defines the interface for the Pub/Sub facade's publish
// side. Concrete implementations are obtained from NewPubSub; the
// generic free function Publish[T] is the recommended caller-facing
// API.
type Publisher interface {
	publish(ctx context.Context, t reflect.Type, payload any) error
}

// Subscriber defines the interface for the Pub/Sub facade's subscribe
// side. Concrete implementations are obtained from NewPubSub; the
// generic free function Subscribe[T] is the recommended caller-facing
// API.
type Subscriber interface {
	subscribe(t reflect.Type, handler func(ctx context.Context, payload any) error) (Cancel, error)
}

// NewPubSub creates a PubSub facade. By default it uses a
// PipelineChannel for every Go type; pass WithChannelFactory to override.
func NewPubSub(opts ...PubSubOption) *PubSub {
	options := &pubSubOptions{
		factory: defaultChannelFactory,
	}

	for _, opt := range opts {
		opt(options)
	}

	return &PubSub{
		factory:  options.factory,
		channels: map[reflect.Type]Channel[any]{},
	}
}

// PubSubOption is a functional option for configuring NewPubSub.
type PubSubOption func(opts *pubSubOptions)

// pubSubOptions holds configuration for the PubSub facade.
type pubSubOptions struct {
	factory ChannelFactory
}

// WithChannelFactory overrides the default channel factory. Nil
// values are silently ignored, preserving the default.
func WithChannelFactory(factory ChannelFactory) PubSubOption {
	return func(opts *pubSubOptions) {
		if factory != nil {
			opts.factory = factory
		}
	}
}

// defaultChannelFactory constructs a PipelineChannel[any] for any Go
// type. Type isolation is preserved because the PubSub registry keys
// each channel by reflect.Type — a handler subscribed for type X
// never observes payloads published for type Y.
func defaultChannelFactory(_ reflect.Type) Channel[any] {
	return NewPipelineChannel[any]()
}

// publish dispatches payload to the channel registered for the given
// type. Implements Publisher.
func (p *PubSub) publish(ctx context.Context, t reflect.Type, payload any) error {
	cassert.NotNil(p, "PubSub is nil")

	channel := p.channelFor(t)

	msg := Message[any]{Payload: payload}

	return channel.Send(ctx, msg)
}

// subscribe registers handler on the channel for the given type.
// Implements Subscriber.
func (p *PubSub) subscribe(t reflect.Type, handler func(ctx context.Context, payload any) error) (Cancel, error) {
	cassert.NotNil(p, "PubSub is nil")

	channel := p.channelFor(t)

	wrapped := func(ctx context.Context, msg Message[any]) error {
		return handler(ctx, msg.Payload)
	}

	return channel.Subscribe(wrapped)
}

// channelFor returns the channel for the given type, allocating it
// lazily via the configured factory.
func (p *PubSub) channelFor(t reflect.Type) Channel[any] {
	p.mu.Lock()
	defer p.mu.Unlock()

	channel, ok := p.channels[t]
	if ok {
		return channel
	}

	channel = p.factory(t)
	p.channels[t] = channel

	return channel
}

// Publish sends payload through pub, routed by the Go type of T.
// Handlers subscribed via Subscribe[T] receive the payload; handlers
// subscribed for other types are not invoked.
func Publish[T any](ctx context.Context, pub Publisher, payload T) error {
	if pub == nil {
		return ErrSend(ErrClosed)
	}

	t := reflect.TypeFor[T]()

	return pub.publish(ctx, t, payload)
}

// Subscribe registers handler for payloads of type T on sub. The
// returned Cancel detaches the handler.
func Subscribe[T any](sub Subscriber, handler Handler[T]) (Cancel, error) {
	if sub == nil {
		return nil, ErrSubscribe(ErrClosed)
	}

	if handler == nil {
		return nil, ErrSubscribe(ErrHandlerNil)
	}

	t := reflect.TypeFor[T]()

	wrapped := func(ctx context.Context, payload any) error {
		typed, ok := payload.(T)
		if !ok {
			// Defensive: same reflect.Type guarantees the assertion
			// succeeds. Wrong-typed payload reaching here means
			// caller corrupted the registry — return nil to keep
			// pub/sub flowing.
			return nil
		}

		return handler(ctx, Message[T]{Payload: typed})
	}

	return sub.subscribe(t, wrapped)
}
