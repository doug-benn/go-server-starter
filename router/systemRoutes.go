package router

import (
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func HandleGetHealth() http.HandlerFunc {
	type responseBody struct {
		Version        string    `json:"Version"`
		Uptime         string    `json:"Uptime"`
		LastCommitHash string    `json:"LastCommitHash"`
		LastCommitTime time.Time `json:"LastCommitTime"`
		DirtyBuild     bool      `json:"DirtyBuild"`
	}

	res := responseBody{Version: "0.1"}
	buildInfo, _ := debug.ReadBuildInfo()
	for _, kv := range buildInfo.Settings {
		if kv.Value == "" {
			continue
		}
		switch kv.Key {
		case "vcs.revision":
			res.LastCommitHash = kv.Value
		case "vcs.time":
			res.LastCommitTime, _ = time.Parse(time.RFC3339, kv.Value)
		case "vcs.modified":
			res.DirtyBuild = kv.Value == "true"
		}
	}

	up := time.Now()
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		res.Uptime = time.Since(up).String()
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// handleGetDebug returns an [http.Handler] for debug routes, including pprof and expvar routes.
func HandleGetDebug() http.Handler {
	mux := http.NewServeMux()

	// NOTE: this route is same as defined in net/http/pprof init function
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// NOTE: this route is same as defined in expvar init function
	mux.Handle("/debug/vars", expvar.Handler())
	return mux
}

// recovery is a middleware that recovers from panics during HTTP handler execution and logs the error details.
// It must be the last middleware in the chain to ensure it captures all panics.
func Recovery(next http.Handler, logger zerolog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {

			err := recover()
			if err == nil {
				return
			}

			if err, ok := err.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				// Handle the abort gracefully
				return
			}

			stack := make([]byte, 1024)
			n := runtime.Stack(stack, true)

			logger.Error().
				Ctx(r.Context()).
				Any("panic!", err).
				Str("stack", string(stack[:n])).
				Str("stack", string(stack[:n])).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("query", r.URL.RawQuery).
				Str("ip", r.RemoteAddr).
				Msg("panic!")

			// send error response
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		}()
		next.ServeHTTP(w, r)
	})
}

func Accesslog(next http.Handler, logger zerolog.Logger) http.Handler {

	h := hlog.NewHandler(logger)

	accessHandler := hlog.AccessHandler(
		func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				//Stringer("url", r.URL).
				Str("path", r.URL.Path).
				Str("query", r.URL.RawQuery).
				Int("status_code", status).
				Int("size_bytes", size).
				Dur("elapsed_ms", duration).
				Msg("completed request")

		},
	)

	//userAgentHandler := hlog.UserAgentHandler("http_user_agent")
	remoteAddrHandler := hlog.RemoteAddrHandler("remote ip")

	return h(accessHandler(remoteAddrHandler(next)))
}

// func Logger(logger zerolog.Logger) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		fn := func(rw http.ResponseWriter, r *http.Request) {
// 			ww := middleware.NewWrapResponseWriter(rw, r.ProtoMajor)
// 			start := time.Now()
// 			defer func() {
// 				logger.Info().
// 					Str("request-id", GetReqID(r.Context())).
// 					Int("status", ww.Status()).
// 					Int("bytes", ww.BytesWritten()).
// 					Str("method", r.Method).
// 					Str("path", r.URL.Path).
// 					Str("query", r.URL.RawQuery).
// 					Str("ip", r.RemoteAddr).
// 					Str("trace.id", trace.SpanFromContext(r.Context()).SpanContext().TraceID().String()).
// 					Str("user-agent", r.UserAgent()).
// 					Dur("latency", time.Since(start)).
// 					Msg("request completed")
// 			}()

// 			next.ServeHTTP(ww, r)
// 		}
// 		return http.HandlerFunc(fn)
// 	}
// }
