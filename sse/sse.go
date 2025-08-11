// Package sse provides utilities for working with Server Sent Events (SSE).
package sse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/doug-benn/go-server-starter/producer"
)

// SSE errors
var (
	ErrHTTPFlush = errors.New("error: unable to flush")
)

// WriteTimeout is the timeout for writing to the client.
var WriteTimeout = 5 * time.Second

type unwrapper interface {
	Unwrap() http.ResponseWriter
}

type writeDeadliner interface {
	SetWriteDeadline(time.Time) error
}

type Event struct {
	ID    int
	Type  string
	Data  any
	Retry int
}

// SSEHandler creates an HTTP handler that serves Server-Sent Events using the producer
func SSEHandler[T Event](producer *producer.Producer[Event], logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Get the flusher/deadliner from the response writer if possible.
		var flusher http.Flusher
		flushCheck := w
		for {
			if f, ok := flushCheck.(http.Flusher); ok {
				flusher = f
				break
			}
			if u, ok := flushCheck.(unwrapper); ok {
				flushCheck = u.Unwrap()
			} else {
				break
			}
		}

		var deadliner writeDeadliner
		deadlineCheck := w
		for {
			if d, ok := deadlineCheck.(writeDeadliner); ok {
				deadliner = d
				break
			}
			if u, ok := deadlineCheck.(unwrapper); ok {
				deadlineCheck = u.Unwrap()
			} else {
				break
			}
		}

		// Subscribe to the producer with a reasonable buffer size
		// Adjust buffer size based on your expected event rate
		subscription := producer.Subscribe(100)

		// Send initial connection message
		fmt.Fprintf(w, "data: {\"type\":\"connected\",\"timestamp\":\"%s\"}\n\n", time.Now().Format(time.RFC3339))
		flusher.Flush()

		// Create context that cancels when client disconnects
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// Listen for events and send them to the client
		for {
			event, err := subscription.Next(ctx)
			if err != nil {
				// Client disconnected or context cancelled
				break
			}

			if deadliner != nil {
				if err := deadliner.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
					logger.Error("unable to set write deadline", "error", err)
				}
			} else {
				logger.Warn("write deadline not supported by underlying writer")
			}

			// Write optional fields
			if event.ID > 0 {
				w.Write(fmt.Appendf(nil, "id: %d\n", event.ID))
			}
			if event.Retry > 0 {
				w.Write(fmt.Appendf(nil, "retry: %d\n", event.Retry))
			}

			if event.Type != "" && event.Type != "message" {
				// `message` is the default, so no need to transmit it.
				w.Write([]byte("event: " + event.Type + "\n"))
			}

			// Write the message data.
			if _, err := w.Write([]byte("data: ")); err != nil {
				w.Write([]byte(`{"error": "encode error: `))
				w.Write([]byte(err.Error()))
				w.Write([]byte("\"}\n\n"))
			}
			encoder := json.NewEncoder(w)
			if err := encoder.Encode(event.Data); err != nil {
				w.Write([]byte(`{"error": "encode error: `))
				w.Write([]byte(err.Error()))
				w.Write([]byte("\"}\n\n"))
			}
			w.Write([]byte("\n"))
			if flusher != nil {
				flusher.Flush()
			} else {
				logger.Error("unable to flush", "error", err)
			}

			// Check if client disconnected
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}
}
