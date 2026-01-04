package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestChainExecutionOrder verifies middleware executes in correct onion pattern
func TestChainExecutionOrder(t *testing.T) {
	// Create tracking middleware to verify execution order
	var executionOrder []string

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "middleware1-before")
			next.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "middleware1-after")
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "middleware2-before")
			next.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "middleware2-after")
		})
	}

	middleware3 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "middleware3-before")
			next.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "middleware3-after")
		})
	}

	// Create chain in order: middleware3, middleware2, middleware1
	chain := NewChain(middleware3, middleware2, middleware1)

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		executionOrder = append(executionOrder, "handler")
		w.WriteHeader(200)
		w.Write([]byte("success"))
	})

	wrappedHandler := chain.Build(finalHandler)

	// Execute request
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	// Verify execution order: should be onion-like (outermost first)
	expectedOrder := []string{
		"middleware3-before", // outermost (last in chain)
		"middleware2-before",
		"middleware1-before",
		"handler",
		"middleware1-after",
		"middleware2-after",
		"middleware3-after", // outermost (last in chain)
	}

	assert.Equal(t, expectedOrder, executionOrder, "Middleware execution order should follow onion pattern")
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "success", rr.Body.String())
}

// TestChainBuildWithNilHandler tests behavior with nil handler
func TestChainBuildWithNilHandler(t *testing.T) {
	chain := NewChain(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write([]byte("middleware-only"))
			})
		},
	)

	handler := chain.Build(nil) // Should create http.NewServeMux()

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "middleware-only", rr.Body.String())
}

// TestChainBuildEmpty tests behavior with empty middleware chain
func TestChainBuildEmpty(t *testing.T) {
	chain := NewChain()

	originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("original"))
	})

	wrappedHandler := chain.Build(originalHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "original", rr.Body.String())
}

// TestChainMiddlewarePanicHandling tests panic propagation through chain
func TestChainMiddlewarePanicHandling(t *testing.T) {
	var panicCaught bool
	var panicMessage interface{}

	panicMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					panicCaught = true
					panicMessage = err
					http.Error(w, "panic handled", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	chain := NewChain(panicMiddleware)
	wrappedHandler := chain.Build(panicHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		wrappedHandler.ServeHTTP(rr, req)
	})

	assert.True(t, panicCaught)
	assert.Equal(t, "test panic", panicMessage)
	assert.Equal(t, 500, rr.Code)
	assert.Equal(t, "panic handled", rr.Body.String())
}

// TestChainSingleMiddleware tests chain with only one middleware
func TestChainSingleMiddleware(t *testing.T) {
	var middlewareExecuted bool

	singleMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareExecuted = true
			next.ServeHTTP(w, r)
		})
	}

	chain := NewChain(singleMiddleware)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("response"))
	})

	wrappedHandler := chain.Build(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	assert.True(t, middlewareExecuted)
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "response", rr.Body.String())
}

// TestChainMultipleRequests tests that middleware chain works across multiple requests
func TestChainMultipleRequests(t *testing.T) {
	var requestCount int

	countingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			next.ServeHTTP(w, r)
		})
	}

	chain := NewChain(countingMiddleware)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	wrappedHandler := chain.Build(handler)

	// Execute multiple requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)

		assert.Equal(t, 200, rr.Code)
		assert.Equal(t, "ok", rr.Body.String())
	}

	assert.Equal(t, 3, requestCount, "Middleware should execute for each request")
}
