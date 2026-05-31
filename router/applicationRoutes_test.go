package router

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
)

func TestHandleHelloWorld_RequestCounter(t *testing.T) {
	c := cache.New(5*time.Minute, 10*time.Minute)
	handler := HandleHelloWorld(slog.Default(), c)

	req := httptest.NewRequest(http.MethodGet, "/helloworld", nil)

	for want := 1; want <= 3; want++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["message"] != "Hello World" {
			t.Errorf("expected message 'Hello World', got %v", resp["message"])
		}

		count, ok := resp["count"].(float64)
		if !ok || int(count) != want {
			t.Errorf("expected count %d, got %v", want, resp["count"])
		}
	}
}
