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

	"github.com/doug-benn/go-json-api/database"
	"github.com/doug-benn/go-json-api/router"
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

	// var port uint
	// fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	// fs.SetOutput(w)
	// fs.UintVar(&port, "port", 8080, "port for HTTP API")
	// if err := fs.Parse(args[1:]); err != nil {
	// 	return err
	// }

	port := 8080

	slog.SetDefault(slog.New(slog.NewJSONHandler(w, nil)))

	db, err := database.NewDatabase(slog.Default(), true, true)
	if err != nil {
		slog.Error("%s", err)
	}
	defer db.Stop()
	db.Start(ctx)

	slog.InfoContext(ctx, "database connection started")

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           route(slog.Default(), version),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, "server started", slog.Uint64("port", uint64(port)), slog.String("version", version))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		db.Stop()
		slog.InfoContext(ctx, "shutting down server")
	}

	ctx, cancel = context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()

	return server.Shutdown(ctx)
}

func route(log *slog.Logger, version string) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /helloworld", router.HandleHelloWorld())

	// System Routes for debug/logging
	mux.Handle("GET /health", router.HandleGetHealth(version))
	mux.Handle("/debug/", router.HandleGetDebug())

	handler := router.Accesslog(mux, log)
	handler = router.Recovery(handler, log)
	return handler
}
