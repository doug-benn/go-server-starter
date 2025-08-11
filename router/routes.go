package router

import (
	"log/slog"
	"net/http"

	"github.com/doug-benn/go-server-starter/producer"
	"github.com/doug-benn/go-server-starter/services"
	"github.com/doug-benn/go-server-starter/sse"
	"github.com/patrickmn/go-cache"
)

func AddRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	cache *cache.Cache,
	producer *producer.Producer[sse.Event],
	todoService services.TodoService,
) {

	//Register all routes
	mux.Handle("GET /helloworld", HandleHelloWorld(logger, cache))
	mux.Handle("GET /todos", HandleGetTodos(logger, todoService))
	mux.Handle("/events", sse.SSEHandler(producer, logger))

	// System Routes for debugging
	mux.Handle("GET /health", HandleGetHealth())
	mux.Handle("/debug/", HandleGetDebug())
	mux.Handle("/", http.NotFoundHandler())
}
