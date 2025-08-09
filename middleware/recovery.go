package middleware

import (
	"log/slog"
	"net/http"
	"runtime"
)

func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					buf := make([]byte, 2048)
					n := runtime.Stack(buf, false)
					buf = buf[:n]

					logger.ErrorContext(r.Context(),
						"panic!",
						slog.Any("error", err),
						slog.String("stack", string(buf)),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.String("query", r.URL.RawQuery),
						slog.String("ip", r.RemoteAddr),
					)

					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
