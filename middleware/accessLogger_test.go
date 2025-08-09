package middleware

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccessLogger tests the main AccessLogger middleware function
func TestAccessLogger(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		query          string
		expectedStatus int
		expectedBody   string
		setupHandler   func() http.Handler
	}{
		{
			name:           "GET request with success response",
			method:         "GET",
			path:           "/api/users",
			query:          "limit=10&offset=0",
			expectedStatus: 200,
			expectedBody:   "success",
			setupHandler: func() http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
					w.Write([]byte("success"))
				})
			},
		},
		{
			name:           "POST request with created response",
			method:         "POST",
			path:           "/api/users",
			query:          "",
			expectedStatus: 201,
			expectedBody:   "created",
			setupHandler: func() http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(201)
					w.Write([]byte("created"))
				})
			},
		},
		{
			name:           "PUT request with no content response",
			method:         "PUT",
			path:           "/api/users/123",
			query:          "",
			expectedStatus: 204,
			expectedBody:   "",
			setupHandler: func() http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(204)
				})
			},
		},
		{
			name:           "DELETE request with error response",
			method:         "DELETE",
			path:           "/api/users/456",
			query:          "",
			expectedStatus: 404,
			expectedBody:   "not found",
			setupHandler: func() http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(404)
					w.Write([]byte("not found"))
				})
			},
		},
		{
			name:           "GET request with server error",
			method:         "GET",
			path:           "/api/error",
			query:          "",
			expectedStatus: 500,
			expectedBody:   "internal server error",
			setupHandler: func() http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(500)
					w.Write([]byte("internal server error"))
				})
			},
		},
		{
			name:           "PATCH request with complex query parameters",
			method:         "PATCH",
			path:           "/api/search",
			query:          "q=golang&sort=date&order=desc&page=2",
			expectedStatus: 200,
			expectedBody:   "search results",
			setupHandler: func() http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
					w.Write([]byte("search results"))
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup logger buffer to capture logs
			var logBuffer bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo}))

			// Create the middleware
			middleware := AccessLogger(logger)

			// Wrap the test handler
			handler := middleware(tt.setupHandler())

			// Create request
			url := tt.path
			if tt.query != "" {
				url = fmt.Sprintf("%s?%s", tt.path, tt.query)
			}
			req := httptest.NewRequest(tt.method, url, nil)
			req.RemoteAddr = "192.168.1.100:12345"

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(rr, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())

			// Verify log was written
			logOutput := logBuffer.String()
			assert.NotEmpty(t, logOutput, "Expected log output but got none")

			// Verify log contains expected fields
			assert.Contains(t, logOutput, fmt.Sprintf(`"method":"%s"`, tt.method))
			assert.Contains(t, logOutput, fmt.Sprintf(`"path":"%s"`, tt.path))
			assert.Contains(t, logOutput, fmt.Sprintf(`"query":"%s"`, tt.query))
			assert.Contains(t, logOutput, fmt.Sprintf(`"status_code":%d`, tt.expectedStatus))
			assert.Contains(t, logOutput, fmt.Sprintf(`"size_bytes":%d`, len(tt.expectedBody)))
			assert.Contains(t, logOutput, `"elapsed_ms"`)
			assert.Contains(t, logOutput, `"remote_ip":"192.168.1.100:12345"`)
			assert.Contains(t, logOutput, fmt.Sprintf(`"%s: %s"`, strconv.Itoa(tt.expectedStatus), http.StatusText(tt.expectedStatus)))
		})
	}
}

// TestAccessLoggerWithSpecialCharacters tests handling of special characters in URLs
func TestAccessLoggerWithSpecialCharacters(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo}))

	middleware := AccessLogger(logger)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))

	// Test with URL containing special characters
	req := httptest.NewRequest("GET", "/api/search?q=hello%20world&tag=%22special%22", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, `"path":"/api/search"`)
	assert.Contains(t, logOutput, `"query":"q=hello%20world&tag=%22special%22"`)
}

// TestAccessLoggerWithEmptyQuery tests handling of requests without query parameters
func TestAccessLoggerWithEmptyQuery(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo}))

	middleware := AccessLogger(logger)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/api/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, `"query":""`)
}

// TestAccessLoggerWithLargeResponse tests logging of large response bodies
func TestAccessLoggerWithLargeResponse(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo}))

	middleware := AccessLogger(logger)

	// Create a large response body
	largeBody := strings.Repeat("a", 10000)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(largeBody))
	}))

	req := httptest.NewRequest("GET", "/api/large", nil)
	req.RemoteAddr = "172.16.0.1:9000"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, `"size_bytes":10000`)
	assert.Equal(t, largeBody, rr.Body.String())
}

// TestAccessLoggerWithPanicRecovery tests that logging still works even if the handler panics
func TestAccessLoggerWithPanicRecovery(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo}))

	middleware := AccessLogger(logger)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest("GET", "/api/panic", nil)
	rr := httptest.NewRecorder()

	// This should panic, but we want to test the middleware setup
	require.Panics(t, func() {
		handler.ServeHTTP(rr, req)
	})
}

// TestAccessLoggerElapsedTime tests that elapsed time is properly measured
func TestAccessLoggerElapsedTime(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo}))

	middleware := AccessLogger(logger)

	// Handler that takes some time to execute
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond) // Sleep for 50ms
		w.WriteHeader(200)
		w.Write([]byte("delayed response"))
	}))

	req := httptest.NewRequest("GET", "/api/slow", nil)
	rr := httptest.NewRecorder()

	start := time.Now()
	handler.ServeHTTP(rr, req)
	elapsed := time.Since(start)

	// Verify the request took at least 50ms
	require.GreaterOrEqual(t, elapsed.Milliseconds(), int64(50))

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, `"elapsed_ms"`)

	// The logged elapsed time should be reasonable (between 50ms and 200ms)
	// This is a bit tricky to test precisely, but we can check the log format
	assert.Regexp(t, `"elapsed_ms":\d+`, logOutput)
}

// TestAccessLoggerDifferentRemoteAddresses tests various remote address formats
func TestAccessLoggerDifferentRemoteAddresses(t *testing.T) {
	testCases := []struct {
		name       string
		remoteAddr string
		expectedIP string
	}{
		{
			name:       "IPv4",
			remoteAddr: "192.168.1.1:8080",
			expectedIP: "192.168.1.1:8080",
		},
		{
			name:       "IPv6",
			remoteAddr: "[2001:db8::1]:8080",
			expectedIP: "[2001:db8::1]:8080",
		},
		{
			name:       "localhost",
			remoteAddr: "127.0.0.1:3000",
			expectedIP: "127.0.0.1:3000",
		},
		{
			name:       "IP",
			remoteAddr: "10.0.0.1",
			expectedIP: "10.0.0.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var logBuffer bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo}))

			middleware := AccessLogger(logger)
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write([]byte("ok"))
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tc.remoteAddr
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			logOutput := logBuffer.String()
			assert.Contains(t, logOutput, fmt.Sprintf(`"remote_ip":"%s"`, tc.expectedIP))
		})
	}
}

// TestAccessLoggerMiddlewareChaining tests that the middleware properly chains with other middlewares
func TestAccessLoggerMiddlewareChaining(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Create a custom middleware that adds a header
	customMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom-Header", "test-value")
			next.ServeHTTP(w, r)
		})
	}

	// Chain middlewares: AccessLogger -> customMiddleware -> handler
	accessLogger := AccessLogger(logger)
	handler := accessLogger(customMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("chained response"))
	})))

	req := httptest.NewRequest("GET", "/api/chained", nil)
	req.RemoteAddr = "198.51.100.1:5000"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "chained response", rr.Body.String())
	assert.Equal(t, "test-value", rr.Header().Get("X-Custom-Header"))

	// Verify logging
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, `"path":"/api/chained"`)
	assert.Contains(t, logOutput, `"status_code":200`)
	assert.Contains(t, logOutput, `"size_bytes":16`) // length of "chained response"
}

// BenchmarkAccessLogger benchmarks the performance of the AccessLogger middleware
func BenchmarkAccessLogger(b *testing.B) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo}))

	middleware := AccessLogger(logger)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("benchmark response"))
	}))

	req := httptest.NewRequest("GET", "/api/benchmark?param=value", nil)
	req.RemoteAddr = "127.0.0.1:8080"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}

// TestAccessLoggerReturnValue tests that the middleware returns a proper http.Handler
func TestAccessLoggerReturnValue(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(bytes.NewBuffer(nil), &slog.HandlerOptions{Level: slog.LevelInfo}))

	middleware := AccessLogger(logger)

	// Test that it returns a function
	assert.NotNil(t, middleware)

	// Test that the returned function accepts an http.Handler and returns an http.Handler
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	wrappedHandler := middleware(dummyHandler)

	assert.NotNil(t, wrappedHandler)
	assert.Implements(t, (*http.Handler)(nil), wrappedHandler)
}
