package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	metricsware "github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"

	"github.com/doug-benn/go-server-starter/database"
	"github.com/doug-benn/go-server-starter/middleware"
	"github.com/doug-benn/go-server-starter/producer"
	"github.com/doug-benn/go-server-starter/router"
	"github.com/doug-benn/go-server-starter/sse"
	"github.com/patrickmn/go-cache"
)

func main() {
	if err := run(os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}
	fmt.Printf("Loaded Config: %+v\n", config)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cache := cache.New(5*time.Minute, 10*time.Minute)

	// Database Connection
	postgresDatabase, err := database.NewDatabase(ctx, logger, database.DefaultConfig())
	if err != nil {
		logger.Error("error creating database pool on startup", "error", err)
		panic(err)
	}

	// Create a producer for FizzBuzz events with a 5-second broadcast timeout
	fizzBuzzProducer := producer.NewProducer(
		producer.WithBroadcastTimeout[sse.Event](5 * time.Second),
	)

	// Start the producer in a goroutine
	go fizzBuzzProducer.Start(ctx)

	// Producer goroutine - generates FizzBuzz events
	go func() {

		for i := 1; true; i = (i + 1) % 1000 {
			var result string

			switch {
			case i%15 == 0:
				result = "FizzBuzz"
			case i%3 == 0:
				result = "Fizz"
			case i%5 == 0:
				result = "Buzz"
			default:
				result = fmt.Sprintf("%d", i)
			}

			event := sse.Event{
				Type: "FixxBuzz Event",
				Data: result,
			}

			fizzBuzzProducer.Broadcast(ctx, event)

			// Add some delay between broadcasts
			time.Sleep(100 * time.Millisecond)
		}

	}()

	mux := http.NewServeMux()
	router.AddRoutes(mux, logger, cache, fizzBuzzProducer)

	handler := middleware.Recovery(logger)(mux)
	handler = middleware.AccessLogger(logger, middleware.IgnorePath("/events"))(mux)

	std.HandlerProvider("", metricsware.New(metricsware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	}))

	// HTTP Server
	server := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", config.Port),
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	metrics := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", 9201), Handler: promhttp.Handler()}

	errChan := make(chan error, 1)

	//Main HTTP Server
	go func() {
		logger.Info("server started", "port", 9200)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	//Metrics Server
	go func() {
		logger.Info("metrics started", "port", 9201)
		if err := metrics.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		logger.InfoContext(ctx, "shutting down server")

		// Create a new context for shutdown with timeout
		ctx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		// Shutdown the HTTP server first
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("HTTP server shutdown: %w", err)
		}

		// Shutdown the metrics server
		if err := metrics.Shutdown(ctx); err != nil {
			return fmt.Errorf("metrics server shutdown: %w", err)
		}

		// cancel the main context
		cancel()

		// services cleanup
		postgresDatabase.Close()

		return nil
	}
}
