package producer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSubscribe(t *testing.T) {
	producer := NewProducer[int]()
	sub := producer.Subscribe(10)
	require.Equal(t, 1, len(producer.subs))
	require.NotNil(t, sub)
}

func TestBroadcast(t *testing.T) {
	producer := NewProducer[int]()
	sub := producer.Subscribe(0)
	done := make(chan bool)
	go func() {
		event, err := sub.Next(context.Background())
		require.NoError(t, err)
		require.Equal(t, 42, event)
		done <- true
	}()
	ctx := context.Background()
	producer.Broadcast(ctx, 42)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out waiting for event")
	}
}

func TestBroadcastTimeout(t *testing.T) {
	timeout := 50 * time.Millisecond
	producer := NewProducer(WithBroadcastTimeout[int](timeout))
	sub := producer.Subscribe(0)

	go func() {
		// Delay sending to simulate timeout scenario
		time.Sleep(100 * time.Millisecond)
		sub.events <- 42
	}()

	event, err := sub.Next(context.Background())
	require.NoError(t, err)
	require.Equal(t, 42, event)
}

func TestEventProducer_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	producer := NewProducer[int]()
	go producer.Start(ctx)

	sub := producer.Subscribe(100)

	// Simulate removing the subscription.
	cancel()
	_, err := sub.Next(ctx)
	if err == nil {
		t.Error("Expected to end after context cancellation")
	}
}

func TestContextCancellationCleansUpBlockingNext(t *testing.T) {
	// Create a producer
	producer := NewProducer[string]()

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Start the producer in a goroutine
	go producer.Start(ctx)

	// Subscribe to events
	sub := producer.Subscribe(10)

	// Channel to track when Next() returns
	nextDone := make(chan error, 1)

	// Start a blocking Next() call in a goroutine
	go func() {
		_, err := sub.Next(context.Background()) // Using background context, not the producer's context
		nextDone <- err
	}()

	// Give the goroutine time to start and block on Next()
	time.Sleep(10 * time.Millisecond)

	// Cancel the producer's context
	cancel()

	// Next() should unblock within a reasonable time
	select {
	case err := <-nextDone:
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled error, got: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Next() did not unblock after producer context was cancelled")
	}
}

func TestContextCancellationCleansUpMultipleSubscribers(t *testing.T) {
	producer := NewProducer[int]()
	ctx, cancel := context.WithCancel(context.Background())

	go producer.Start(ctx)

	// Create multiple subscribers
	numSubs := 10
	subs := make([]*Subscription[int], numSubs)
	nextDone := make(chan error, numSubs)

	for i := 0; i < numSubs; i++ {
		subs[i] = producer.Subscribe(5)

		// Start blocking Next() calls for each subscription
		go func(sub *Subscription[int]) {
			_, err := sub.Next(context.Background())
			nextDone <- err
		}(subs[i])
	}

	// Give goroutines time to start and block
	time.Sleep(10 * time.Millisecond)

	// Cancel the producer's context
	cancel()

	// All Next() calls should unblock
	for i := 0; i < numSubs; i++ {
		select {
		case err := <-nextDone:
			if err != context.Canceled {
				t.Errorf("Subscriber %d: Expected context.Canceled error, got: %v", i, err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Subscriber %d: Next() did not unblock after producer context was cancelled", i)
		}
	}
}
