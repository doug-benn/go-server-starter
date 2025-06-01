package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	mmiddleware "github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"

	"github.com/doug-benn/go-server-starter/database"
	"github.com/doug-benn/go-server-starter/logging"
	"github.com/doug-benn/go-server-starter/router"
	"github.com/patrickmn/go-cache"

	"github.com/doug-benn/go-server-starter/middleware"
)

func main() {
	if err := run(context.Background(), os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := logging.NewZeroLogger(logging.NewLoggingConfig())
	logger.Info().Msg("Zero Logger initialized")

	cache := cache.New(5*time.Minute, 10*time.Minute)

	// Database Connection
	postgresDatabase, err := database.NewDatabase(ctx, logger, database.NewConfig())
	if err != nil {
		logger.Error().Err(err).Msg("database error")
	}

	mux := http.NewServeMux()

	//Register all routes
	mux.Handle("GET /helloworld", router.HandleHelloWorld(logger, cache))
	mux.Handle("/events", router.HandleEvents())

	// System Routes for debugging
	mux.Handle("GET /health", router.HandleGetHealth())
	mux.Handle("/debug/", router.HandleGetDebug())

	//middleware chain
	chain := middleware.New(
		middleware.Recovery(logger),
		middleware.AccessLogger(logger),
		std.HandlerProvider("", mmiddleware.New(mmiddleware.Config{
			Recorder: metrics.NewRecorder(metrics.Config{}),
		}))).Build(mux)

	//Metrics Handler
	//handler = std.Handler("", metricsMiddleware, handler)

	// HTTP Server
	server := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", 9200),
		Handler:      chain,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	metrics := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", 9201), Handler: promhttp.Handler()}

	errChan := make(chan error, 1)

	//Main HTTP Server
	go func() {
		logger.Info().Int("port", 9200).Msg("server started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	//Metrics Server
	go func() {
		logger.Info().Int("port", 9201).Msg("metrics started")
		if err := metrics.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	// Server Shutdown
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		postgresDatabase.Close()
		logger.Info().Msg("shutting down server")
	}

	ctx, cancel = context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}
