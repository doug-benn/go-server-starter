package router

import (
	"net/http"

	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog"
)

func RegisterRoutes(mux *http.ServeMux, logger zerolog.Logger, cache *cache.Cache) {

	mux.Handle("GET /helloworld", HandleHelloWorld(logger, cache))

	// System Routes for debug/logging
	mux.Handle("GET /health", HandleGetHealth())
	mux.Handle("/debug/", HandleGetDebug())

	mux.Handle("/events", HandleEvents())

}
