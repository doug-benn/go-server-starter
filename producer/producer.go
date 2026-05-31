// Credit to: Raul Jordan https://rauljordan.com/no-sleep-until-we-build-the-perfect-library-in-go/
package producer

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"
)

type subId uint64

const (
	defaultBroadcastTimeout = time.Minute
	defaultMaxWorkers      = 32
)

// Producer manages event subscriptions and broadcasts events to them.
type Producer[T any] struct {
	sync.RWMutex
	subs             map[subId]*Subscription[T]
	nextID           subId
	doneListener     chan subId    // channel to listen for IDs of subscriptions to be removed.
	broadcastTimeout time.Duration // maximum duration to wait for an event to be sent.
	maxWorkers       int           // maximum concurrent goroutines per Broadcast call.
	logger           *slog.Logger
}

type ProducerOpt[T any] func(*Producer[T])

// WithBroadcastTimeout enables the amount of time the broadcaster will wait to send
// to each subscriber before dropping the send.
func WithBroadcastTimeout[T any](timeout time.Duration) ProducerOpt[T] {
	return func(ep *Producer[T]) {
		ep.broadcastTimeout = timeout
	}
}

// WithMaxWorkers sets the maximum number of concurrent goroutines per Broadcast call.
func WithMaxWorkers[T any](n int) ProducerOpt[T] {
	return func(ep *Producer[T]) {
		if n > 0 {
			ep.maxWorkers = n
		}
	}
}

func WithCustomLogger[T any](logger *slog.Logger) ProducerOpt[T] {
	return func(ep *Producer[T]) {
		ep.logger = logger
	}
}

func NewProducer[T any](opts ...ProducerOpt[T]) *Producer[T] {
	producer := &Producer[T]{
		subs:             make(map[subId]*Subscription[T]),
		doneListener:     make(chan subId, 100),
		broadcastTimeout: defaultBroadcastTimeout,
		maxWorkers:       defaultMaxWorkers,
		logger:           slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
	for _, opt := range opts {
		opt(producer)
	}
	return producer
}

// Start begins listening for subscription cancelation requests or context cancelation.
func (ep *Producer[T]) Start(ctx context.Context) {
	for {
		select {
		case id := <-ep.doneListener:
			ep.Lock()
			if sub, exists := ep.subs[id]; exists {
				close(sub.events)
				delete(ep.subs, id)
			}
			ep.Unlock()
		case <-ctx.Done():
			ep.logger.Info("context cancelled producer closing")
			// Clean up all subscriptions
			ep.Lock()
			for _, sub := range ep.subs {
				close(sub.events)
			}
			// Clear the map
			ep.subs = make(map[subId]*Subscription[T])
			ep.Unlock()

			close(ep.doneListener)
			return
		}
	}
}

// Subscribe to events emitted by the producer with some buffer size.
// If the subscriber consumes and processes events slower than what the producer emits,
// there is a chance the producer can drop the event if the subscriber takes longer than the DEFAULT_BROADCAST_TIMEOUT
// duration. It is recommended to specify a buffer size to ensure event emission does not get blocked and that subscribers
// always receive their required events.
//
// To compute an optimal buffer size for a channel given the event production rate 𝑃 and
// consumption rate 𝑄, consider the following:
// If production is faster than consumption, buffer size needs to be large enough to accommodate excess events.
// A basic way to determine a recommended buffer size is (P−Q)×T, where T is a time period over which
// the subscriber needs to handle basic events. If there are 100 events per second, and the processing routine
// can only handle 90, being able to handle excess events over a 10 second period gives us a minimum buffer size of 100.
func (ep *Producer[T]) Subscribe(bufferSize int) *Subscription[T] {
	ep.logger.Info("new subscriber subscribing")
	ep.Lock()
	defer ep.Unlock()
	id := ep.nextID
	ep.nextID++
	sub := &Subscription[T]{
		id:     id,
		events: make(chan T, bufferSize),
		done:   ep.doneListener,
		logger: ep.logger,
	}
	ep.subs[id] = sub
	return sub
}

// Broadcast sends an event to all active subscriptions, respecting a configured timeout or context.
// It spawns goroutines to send events to each subscription so as to not block the producer from
// submitting to all consumers. The number of concurrent goroutines is capped by maxWorkers.
func (ep *Producer[T]) Broadcast(ctx context.Context, event T) {
	ep.RLock()
	subs := make([]*Subscription[T], 0, len(ep.subs))
	for _, sub := range ep.subs {
		subs = append(subs, sub)
	}
	ep.RUnlock()

	sem := make(chan struct{}, ep.maxWorkers)
	var wg sync.WaitGroup
	for _, sub := range subs {
		sem <- struct{}{}
		wg.Add(1)
		go func(listener *Subscription[T]) {
			defer wg.Done()
			defer func() { <-sem }()
			select {
			case listener.events <- event:
			case <-time.After(ep.broadcastTimeout):
				ep.logger.Warn("subscriber too slow, dropping event",
					"subscriber_id", listener.id,
					"buffer_capacity", cap(listener.events),
					"broadcast_timeout", ep.broadcastTimeout,
				)
			case <-ctx.Done():
			}
		}(sub)
	}
	wg.Wait()
}

// Subscription defines a generic handle to a subscription of
// events from a producer.
type Subscription[T any] struct {
	id     subId
	events chan T
	done   chan subId
	logger *slog.Logger
}

// Events returns a read-only channel of events from the subscription.
func (es *Subscription[T]) Events() <-chan T {
	return es.events
}

// Close sends the subscription ID to the producer's done listener for cleanup.
func (es *Subscription[T]) Close() {
	es.done <- es.id
}

// Next waits for the next event or context cancelation, returning the event or an error.
func (es *Subscription[T]) Next(ctx context.Context) (T, error) {
	var zeroVal T
	select {
	case ev, ok := <-es.events:
		if !ok {
			// Channel was closed, producer is shutting down
			return zeroVal, context.Canceled
		}
		return ev, nil
	case <-ctx.Done():
		es.logger.Info("subscriber disconnected")
		es.done <- es.id
		return zeroVal, ctx.Err()
	}
}
