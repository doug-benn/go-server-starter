package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func AccessLogger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		h := hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {

			//TODO Add skippable URL path functionality to access logger
			////Create Skippable URL paths that wont be logged
			// for _, skipPath := range SkipPaths {
			// 	if r.URL.Path == skipPath {
			// 		return
			// 	}
			// }

			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("query", r.URL.RawQuery).
				Int("status_code", status).
				Int("size_bytes", size).
				Dur("elapsed_ms", duration).
				Msg("completed request")
		})(next)

		// Add remote address handler
		h = hlog.RemoteAddrHandler("remote_ip")(h)

		// Add request ID handler
		//h = hlog.RequestIDHandler("req_id", "Request-Id")(h)

		// Add logger
		h = hlog.NewHandler(logger)(h)

		return h
	}
}
