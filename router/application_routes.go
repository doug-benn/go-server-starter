package router

import (
	"encoding/json"
	"net/http"
	"time"
)

func HandleHelloWorld() http.HandlerFunc {
	type responseBody struct {
		Message string `json:"Message"`
		Uptime  string `json:"Uptime"`
	}

	res := responseBody{Message: "Hello World"}

	up := time.Now()
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		res.Uptime = time.Since(up).String()
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
