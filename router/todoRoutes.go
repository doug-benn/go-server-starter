package router

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/doug-benn/go-server-starter/services"
)

func HandleGetTodos(logger *slog.Logger, todoService services.TodoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := todoService.GetAllTodos(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
