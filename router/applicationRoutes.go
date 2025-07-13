package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog"
)

func HandleHelloWorld(logger zerolog.Logger, cache *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		data := "Hello World"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandleSSEEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set http headers required for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// You may need this locally for CORS requests
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Create a channel for client disconnection
		clientGone := r.Context().Done()

		rc := http.NewResponseController(w)
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-clientGone:
				fmt.Println("Client disconnected")
				return
			case <-t.C:
				// Send an event to the client
				// Here we send only the "data" field, but there are few others
				_, err := fmt.Fprintf(w, "data: The time is %s\n\n", time.Now().Format(time.UnixDate))
				if err != nil {
					return
				}
				err = rc.Flush()
				if err != nil {
					fmt.Println("No flushing!")
					return
				}
			}
		}
	}
}
