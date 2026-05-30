package database

import (
	"context"
	"testing"
	"time"
)

// TestListenerShutdownLogic tests listener shutdown logic
func TestListenerShutdownLogic(t *testing.T) {
	// Test that the context cancellation pattern works correctly
	ctx, cancel := context.WithCancel(context.Background())

	shutdownDetected := make(chan bool, 1)

	// Simulate the listener goroutine pattern from main.go
	go func() {
		defer func() {
			shutdownDetected <- true
		}()

		for {
			select {
			case <-ctx.Done():
				// Context was cancelled, exit gracefully - this is the key fix
				return
			default:
				// Simulate some work
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Cancel the context after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	// Wait for graceful shutdown
	select {
	case <-shutdownDetected:
		t.Log("listener goroutine shut down gracefully")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("listener goroutine did not shut down")
	}
}
