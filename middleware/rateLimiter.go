package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func RateLimiter(r rate.Limit, burst int) func(http.Handler) http.Handler {
	var mu sync.Mutex
	clients := make(map[string]*clientLimiter)

	getLimiter := func(key string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()

		now := time.Now()

		if cl, ok := clients[key]; ok {
			cl.lastSeen = now
			return cl.limiter
		}

		limiter := rate.NewLimiter(r, burst)
		clients[key] = &clientLimiter{limiter: limiter, lastSeen: now}

		cleaned := 0
		for k, cl := range clients {
			if now.Sub(cl.lastSeen) > 5*time.Minute {
				delete(clients, k)
				cleaned++
				if cleaned >= 10 {
					break
				}
			}
		}

		return limiter
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.RemoteAddr
			if ff := r.Header.Get("X-Forwarded-For"); ff != "" {
				key = ff
			}

			limiter := getLimiter(key)

			if !limiter.Allow() {
				w.Header().Set("Retry-After", "1")
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(burst))
				http.Error(w, "too many requests\n", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
