package router

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/doug-benn/go-server-starter/services"
	"github.com/patrickmn/go-cache"
)

func HandleHelloWorld(log *slog.Logger, dbService *services.Service, cache *cache.Cache) http.HandlerFunc {
	type responseBody struct {
		Message string `json:"Message"`
		Uptime  string `json:"Uptime"`
	}

	res := responseBody{Message: "Hello World"}

	// ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	// defer cancel()
	//
	up := time.Now()
	return func(w http.ResponseWriter, _ *http.Request) {
		var data responseBody

		if cachedData, found := cache.Get("helloworld"); found {
			fmt.Println("Found Data in Cache")
			data = cachedData.(responseBody)
		} else {
			fmt.Println("Didn't find data is cache, so settings it")
			res.Uptime = time.Since(up).String()
			cache.Set("helloworld", res, 30*time.Second)
			data = res
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
