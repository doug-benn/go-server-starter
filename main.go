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

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"

	"github.com/doug-benn/go-server-starter/database"
	"github.com/doug-benn/go-server-starter/logging"
	"github.com/doug-benn/go-server-starter/router"
	"github.com/patrickmn/go-cache"

	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
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

	logger := logging.NewZeroLogger()
	logger.Info().Msg("Zero Logger initialized")

	cache := cache.New(5*time.Minute, 10*time.Minute)

	// Database Connection
	postgresDatabase, err := database.NewDatabase(ctx, logger)
	if err != nil {
		logger.Error().AnErr("database err", err)
	}
	//postgresDatabase.Stop()

	//Create metrics middleware.
	metricsMiddleware := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	// HTTP Server
	server := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", 9200),
		Handler:      NewApplication(logger, cache, metricsMiddleware),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

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
		if err := http.ListenAndServe("127.0.0.1:9201", promhttp.Handler()); err != nil {
			errChan <- err
		}
	}()

	// Server Shutdown
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		postgresDatabase.Stop()
		slog.InfoContext(ctx, "shutting down server")
		logger.Info().Msg("shutting down server")
	}

	ctx, cancel = context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func NewApplication(logger zerolog.Logger, cache *cache.Cache, metricsMiddleware middleware.Middleware) http.Handler {
	mux := http.NewServeMux()
	router.RegisterRoutes(mux, logger, cache)
	var handler http.Handler = mux

	handler = router.Recovery(handler, logger)
	handler = router.Accesslog(handler, logger)

	//Metrics Handler
	handler = std.Handler("", metricsMiddleware, handler)

	return handler
}
