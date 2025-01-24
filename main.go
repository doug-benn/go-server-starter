package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/doug-benn/go-server-starter/database"
	"github.com/doug-benn/go-server-starter/router"
	"github.com/patrickmn/go-cache"
)

func main() {
	if err := run(context.Background(), os.Stdout, os.Args, Version); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

var Version string

func run(ctx context.Context, w io.Writer, args []string, version string) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// TODO Make this an Env Var
	port := 8080

	slog.SetDefault(slog.New(slog.NewJSONHandler(w, nil)))

	cache := cache.New(5*time.Minute, 10*time.Minute)

	// Database Connection
	dbClient, err := database.NewDatabase(true, true)
	if err != nil {
		slog.Error("%s", err)
	}
	defer dbClient.Stop()
	dbClient.Start(ctx)
	slog.InfoContext(ctx, "database connection started")

	// HTTP Server
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           route(slog.Default(), version, dbClient, cache),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, "server started", slog.Uint64("port", uint64(port)), slog.String("version", version))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Server Shutdown
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		dbClient.Stop()
		slog.InfoContext(ctx, "shutting down server")
	}

	ctx, cancel = context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func corsHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// handle preflight in here
		} else {
			h.ServeHTTP(w, r)
		}
	}
}

func route(log *slog.Logger, version string, dbInterface database.PostgresService, cache *cache.Cache) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /helloworld", router.HandleHelloWorld(log, dbInterface, cache))

	// System Routes for debug/logging
	mux.Handle("GET /health", router.HandleGetHealth(version))
	mux.Handle("/debug/", router.HandleGetDebug())

	handler := router.Accesslog(mux, log)
	handler = router.Recovery(handler, log)
	return handler
}
