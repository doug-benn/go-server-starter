package middleware

import (
	"math"
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

		if cl, ok := clients[key]; ok {
			cl.lastSeen = time.Now()
			return cl.limiter
		}

		limiter := rate.NewLimiter(r, burst)
		clients[key] = &clientLimiter{limiter: limiter, lastSeen: time.Now()}

		if len(clients) > 1000 {
			now := time.Now()
			for k, cl := range clients {
				if now.Sub(cl.lastSeen) > 5*time.Minute {
					delete(clients, k)
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

			reserve := limiter.Reserve()
			delay := reserve.Delay()
			if delay > 0 {
				reserve.Cancel()
				retryAfter := int(math.Ceil(delay.Seconds()))
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(burst))
				http.Error(w, "too many requests\n", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
