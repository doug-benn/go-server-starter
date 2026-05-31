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
		if _, err := fmt.Fprintf(w, "data: {\"type\":\"connected\",\"timestamp\":\"%s\"}\n\n", time.Now().Format(time.RFC3339)); err != nil {
			logger.Error("failed to write connected message", "error", err)
			return
		}
		rc.Flush()

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
			if _, err := w.Write([]byte("data: ")); err != nil {
				w.Write([]byte(`{"error": "encode error: `))
				w.Write([]byte(err.Error()))
				w.Write([]byte("\"}\n\n"))
				logger.Error("failed to write data prefix", "error", err)
			}

			// Handle different data types
			switch data := event.Data.(type) {
			case json.RawMessage: // Already valid JSON bytes - write directly
				if _, err := w.Write(data); err != nil {
					w.Write([]byte(`{"error": "encode error: `))
					w.Write([]byte(err.Error()))
					w.Write([]byte("\"}\n\n"))
					logger.Error("failed to encode raw json", "error", err)
				}
			case []byte: // Treat as JSON RawMessage
				if _, err := w.Write(data); err != nil {
					w.Write([]byte(`{"error": "encode error: `))
					w.Write([]byte(err.Error()))
					w.Write([]byte("\"}\n\n"))
					logger.Error("failed to encode []byte", "error", err)
				}
			case string:
				b, err := json.Marshal(data)
				if err != nil {
					w.Write([]byte(`{"error": "encode error": "`))
					w.Write([]byte(err.Error()))
					w.Write([]byte("\"}\n\n"))
					logger.Error("failed to encode string data", "error", err)
				} else {
					w.Write(b)
				}
			default:
				b, err := json.Marshal(data)
				if err != nil {
					w.Write([]byte(`{"error": "encode error": "`))
					w.Write([]byte(err.Error()))
					w.Write([]byte("\"}\n\n"))
					logger.Error("failed to encode data", "error", err)
				} else {
					w.Write(b)
				}
			}

			w.Write([]byte("\n\n"))

			if err := rc.Flush(); err != nil {
				logger.Error("unable to flush", "error", err)
				return
			}
		}
	}
}
