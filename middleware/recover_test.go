package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecovery(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.Handler
		expectPanic    bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "no panic - normal execution",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			}),
			expectPanic:    false,
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "panic with string",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("something went wrong")
			}),
			expectPanic:    true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error\n",
		},
		{
			name: "panic with error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(http.ErrMissingFile)
			}),
			expectPanic:    true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture log output
			var logBuffer bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo}))

			// Create the recovery middleware
			recoveryMiddleware := Recovery(logger)

			// Wrap the test handler with recovery middleware
			wrappedHandler := recoveryMiddleware(tt.handler)

			// Create test request
			req := httptest.NewRequest("GET", "/test?param=value", nil)
			req = req.WithContext(context.Background())
			req.RemoteAddr = "192.168.1.1:8080"

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute the handler
			wrappedHandler.ServeHTTP(w, req)

			// Check response status
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response body
			if w.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, w.Body.String())
			}

			// Check logging behavior
			logOutput := logBuffer.String()
			if tt.expectPanic {
				// Verify panic was logged
				if !strings.Contains(logOutput, "panic!") {
					t.Error("expected panic to be logged")
				}

				// Verify log contains expected fields
				expectedFields := []string{
					"method", "GET",
					"path", "/test",
					"query", "param=value",
					"ip", "192.168.1.1:8080",
					"stack",
				}

				for _, field := range expectedFields {
					if !strings.Contains(logOutput, field) {
						t.Errorf("expected log to contain %q, got: %s", field, logOutput)
					}
				}
			} else {
				// Verify no panic was logged
				if strings.Contains(logOutput, "panic!") {
					t.Error("unexpected panic logged")
				}
			}
		})
	}
}

func TestRecoveryWithDifferentHTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run("panic_with_"+method, func(t *testing.T) {
			var logBuffer bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo}))

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("test panic")
			})

			recoveryMiddleware := Recovery(logger)
			wrappedHandler := recoveryMiddleware(handler)

			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			if w.Code != http.StatusInternalServerError {
				t.Errorf("expected status 500, got %d", w.Code)
			}

			logOutput := logBuffer.String()
			if !strings.Contains(logOutput, method) {
				t.Errorf("expected log to contain method %s", method)
			}
		})
	}
}

func TestRecoveryPreservesContext(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Create a context with a value
	ctx := context.WithValue(context.Background(), "test-key", "test-value")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context is preserved in the handler
		if r.Context().Value("test-key") != "test-value" {
			t.Error("context not preserved in handler")
		}
		panic("test panic")
	})

	recoveryMiddleware := Recovery(logger)
	wrappedHandler := recoveryMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestRecoveryThroughChain(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Chain order matches main.go: Recovery outermost, then AccessLogger
	chain := NewChain(
		Recovery(logger),
		AccessLogger(logger),
	)

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic from handler")
	})

	wrappedHandler := chain.Build(panicHandler)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "10.0.0.1:8080"
	rr := httptest.NewRecorder()

	// Must NOT panic — Recovery should catch it
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic propagated: %v", r)
		}
	}()
	wrappedHandler.ServeHTTP(rr, req)

	// Response must be 500 with the recovery message
	if rr.Code != 500 {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
	if rr.Body.String() != "internal server error\n" {
		t.Errorf("expected body %q, got %q", "internal server error\n", rr.Body.String())
	}

	logOutput := logBuffer.String()

	// Recovery must log the panic with request context
	if !strings.Contains(logOutput, `"panic!"`) {
		t.Error("expected panic log entry")
	}
	if !strings.Contains(logOutput, `"error":"test panic from handler"`) {
		t.Error("expected panic message in log")
	}
	if !strings.Contains(logOutput, `"method":"GET"`) {
		t.Error("expected method in log")
	}
	if !strings.Contains(logOutput, `"path":"/api/test"`) {
		t.Error("expected path in log")
	}
	if !strings.Contains(logOutput, `"remote_ip":"10.0.0.1:8080"`) {
		t.Error("expected remote IP in log")
	}
	if !strings.Contains(logOutput, `"stack"`) {
		t.Error("expected stack trace in log")
	}

	// AccessLogger must log a request entry (status may be 0 since recovery
	// writes to the original ResponseWriter, not the WriterProxy)
	if !strings.Contains(logOutput, `"status_code":0`) {
		t.Error("expected access log with status 0")
	}
	if !strings.Contains(logOutput, `"size_bytes":0`) {
		t.Error("expected access log with size 0")
	}
}
