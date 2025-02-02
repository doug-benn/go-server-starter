package router

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/doug-benn/go-server-starter/database"
	"github.com/patrickmn/go-cache"
)

func HandleHelloWorld(log *slog.Logger, dbService database.PostgresService, cache *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {

		data, _ := dbService.GetAllComments()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
