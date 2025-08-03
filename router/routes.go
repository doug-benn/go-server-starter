package router

import (
	"net/http"

	"github.com/doug-benn/go-server-starter/producer"
	"github.com/doug-benn/go-server-starter/sse"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog"
)

func AddRoutes(
	mux *http.ServeMux,
	logger zerolog.Logger,
	cache *cache.Cache,
	producer *producer.Producer[sse.Event],
) {

	//Register all routes
	mux.Handle("GET /helloworld", HandleHelloWorld(logger, cache))
	mux.Handle("/events", HandleSSEEvents())

	// System Routes for debugging
	mux.Handle("GET /health", HandleGetHealth())
	mux.Handle("/debug/", HandleGetDebug())
	mux.Handle("/", http.NotFoundHandler())
}
