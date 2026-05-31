package router

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/patrickmn/go-cache"
)

func HandleHelloWorld(logger *slog.Logger, cache *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		data := "Hello World"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
