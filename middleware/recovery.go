package middleware

import (
	"net/http"
	"runtime"

	"github.com/rs/zerolog"
)

func Recovery(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					buf := make([]byte, 2048)
					n := runtime.Stack(buf, false)
					buf = buf[:n]

					logger.Error().
						Ctx(r.Context()).
						Any("panic!", err).
						Str("stack", string(buf)).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Str("query", r.URL.RawQuery).
						Str("ip", r.RemoteAddr).
						Msg("panic!")

					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
