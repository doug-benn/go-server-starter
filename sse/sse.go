// Package sse provides utilities for working with Server Sent Events (SSE).
package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/doug-benn/go-server-starter/producer"
)

// WriteTimeout is the timeout for writing to the client.
const WriteTimeout = 5 * time.Second

// Event represents an SSE event
// Data can be:
// - json.RawMessage ([]byte) - will be written directly as valid JSON
// - string - will be JSON-encoded as a string
// - any other type - will be JSON-encoded
type Event struct {
	ID    int
	Type  string
	Data  any
	Retry int
}

// SSEHandler creates an HTTP handler that serves Server-Sent Events using the producer
func SSEHandler(producer *producer.Producer[Event], logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		rc := http.NewResponseController(w)

		// Subscribe to the producer with a reasonable buffer size
		// Adjust buffer size based on your expected event rate
		subscription := producer.Subscribe(100)

		// Send initial connection message
		if _, err := fmt.Fprintf(w, "event: connected\ndata: {\"timestamp\":\"%s\"}\n\n", time.Now().Format(time.RFC3339)); err != nil {
			logger.Error("failed to write connected message", "error", err)
			return
		}
		rc.Flush()

		// Create context that cancels when client disconnects
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		keepalive := time.NewTicker(25 * time.Second)
		defer keepalive.Stop()

		// Listen for events and send them to the client
		for {
			select {
			case <-ctx.Done():
				subscription.Close()
				return
			case <-keepalive.C:
				fmt.Fprintf(w, "event: keepalive\ndata: {\"timestamp\":\"%s\"}\n\n", time.Now().Format(time.RFC3339))
				rc.Flush()
			case event, ok := <-subscription.Events():
				if !ok {
					return
				}

				if err := rc.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
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
				w.Write([]byte("data: "))

				// Handle different data types
				switch data := event.Data.(type) {
				case json.RawMessage: // Already valid JSON bytes - write directly
					w.Write(data)
				case []byte: // Treat as JSON RawMessage
					w.Write(data)
				case string:
					b, err := json.Marshal(data)
					if err != nil {
						logger.Error("failed to encode string data", "error", err)
						return
					}
					w.Write(b)
				default:
					b, err := json.Marshal(data)
					if err != nil {
						logger.Error("failed to encode data", "error", err)
						return
					}
					w.Write(b)
				}

				w.Write([]byte("\n\n"))

				if err := rc.Flush(); err != nil {
					logger.Error("unable to flush", "error", err)
					return
				}
			}
		}
	}
}
