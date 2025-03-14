package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	sloghttp "github.com/samber/slog-http"

	services "github.com/doug-benn/go-server-starter/Services"
	"github.com/doug-benn/go-server-starter/database"
	"github.com/doug-benn/go-server-starter/logger"
	"github.com/doug-benn/go-server-starter/router"
	"github.com/patrickmn/go-cache"

	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
)

func main() {
	if err := run(context.Background(), os.Stdout, os.Args, Version); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

var Version string

func run(ctx context.Context, w io.Writer, args []string, version string) error {
	//Main Server Context
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// TODO Make this an Env Var? - Tests will needs to be updated
	var port uint
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	fs.SetOutput(w)
	fs.UintVar(&port, "port", 9200, "port for HTTP Server - default 9200")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	slogLogger := logger.New(logger.Config{
		FilePath:         "logs/logs.json",
		UserLocalTime:    false,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 30,
		LogLevel:         slog.LevelInfo,
	}, nil, true)

	zeroLogger := logger.Get()
	zeroLogger.Info().Msg("Zero Logger initialized")

	cache := cache.New(5*time.Minute, 10*time.Minute)

	// Database Connection
	postgresConn, err := database.NewDatabase(true, true)
	if err != nil {
		slogLogger.Error(err.Error())
	}
	postgresConn.Start(ctx)

	todoService := services.NewTaskService(postgresConn.Sql)

	slogLogger.InfoContext(ctx, "database connection started")

	//Create metrics middleware.
	metricsMiddleware := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	// HTTP Server
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           route(slogLogger, version, postgresService, cache, metricsMiddleware),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errChan := make(chan error, 1)
	//Main HTTP Server
	go func() {
		slogLogger.InfoContext(ctx, "server started", slog.Uint64("port", uint64(port)), slog.String("version", version))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	//Metrics Server
	go func() {
		slogLogger.InfoContext(ctx, "metrics listening on", slog.Uint64("port", uint64(port+1)))
		if err := http.ListenAndServe(":9201", promhttp.Handler()); err != nil {
			errChan <- err
		}
	}()

	// Server Shutdown
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		postgresConn.Stop()
		slog.InfoContext(ctx, "shutting down server")
	}

	ctx, cancel = context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func route(logger *slog.Logger, version string, service *services.TaskService, cache *cache.Cache, metricsMiddleware middleware.Middleware) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /helloworld", router.HandleHelloWorld(logger, cache))

	// System Routes for debug/logging
	mux.Handle("GET /health", router.HandleGetHealth(version))
	mux.Handle("/debug/", router.HandleGetDebug())

	mux.Handle("/events", router.HandleEvents())

	handler := sloghttp.Recovery(mux)
	//handler = sloghttp.New(logger)(handler)

	handler = router.RequestLogger(handler)

	//Metrics Handler
	handler = std.Handler("", metricsMiddleware, handler)

	// Wrap main handler with metrics middleware

	return handler
}
