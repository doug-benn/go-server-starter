package router

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/patrickmn/go-cache"
)

func HandleHelloWorld(logger *slog.Logger, c *cache.Cache) http.HandlerFunc {
	c.Add("hello_count", 0, 0)

	return func(w http.ResponseWriter, _ *http.Request) {
		count, err := c.IncrementInt("hello_count", 1)
		if err != nil {
			c.Add("hello_count", 1, 0)
			count = 1
		}

		resp := map[string]any{
			"message": "Hello World",
			"count":   count,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("failed to encode hello world response", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
