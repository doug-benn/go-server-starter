package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/time/rate"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func TestRateLimiterAllowsWithinBurst(t *testing.T) {
	middleware := RateLimiter(rate.Limit(1), 5)
	handler := middleware(okHandler())

	for i := range 5 {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, rr.Code)
		}
	}
}

func TestRateLimiterBlocksExcess(t *testing.T) {
	middleware := RateLimiter(rate.Limit(1), 5)
	handler := middleware(okHandler())

	for range 5 {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rr.Code)
	}

	if rr.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header")
	}

	if rr.Header().Get("X-RateLimit-Limit") != "5" {
		t.Errorf("expected X-RateLimit-Limit: 5, got %s", rr.Header().Get("X-RateLimit-Limit"))
	}
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	middleware := RateLimiter(rate.Limit(1), 3)
	handler := middleware(okHandler())

	clientA := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.1")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr
	}

	clientB := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.2")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr
	}

	// Exhaust client A's burst
	for range 3 {
		rr := clientA()
		if rr.Code != http.StatusOK {
			t.Fatal("client A should be allowed up to burst")
		}
	}

	// Client A should now be rate limited
	rr := clientA()
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("client A expected 429, got %d", rr.Code)
	}

	// Client B should still be allowed (independent bucket)
	rr = clientB()
	if rr.Code != http.StatusOK {
		t.Errorf("client B expected 200, got %d", rr.Code)
	}
}
