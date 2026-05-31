package router

import (
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"runtime/debug"
	"sync"
	"time"

	"github.com/doug-benn/go-server-starter/utilities"
	"github.com/grafana/pyroscope-go"
	pyroscope_pprof "github.com/grafana/pyroscope-go/http/pprof"
)

var pyroscopeOnce sync.Once

func HandleGetHealth() http.HandlerFunc {
	type responseBody struct {
		Version        string    `json:"version"`
		Uptime         string    `json:"uptime"`
		LastCommitHash string    `json:"last_commit_hash"`
		LastCommitTime time.Time `json:"last_commit_time"`
		DirtyBuild     bool      `json:"dirty_build"`
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

// HandleGetDebug returns a handler for debug and profiling endpoints.
// Pyroscope is started lazily on the first request to /debug/pprof/profile.
func HandleGetDebug() http.Handler {
	mux := http.NewServeMux()

	// Standard pprof routes
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Profile route starts Pyroscope on first use
	mux.HandleFunc("/debug/pprof/profile", func(w http.ResponseWriter, r *http.Request) {
		// Extend deadline since CPU profiles typically exceed the server's WriteTimeout
		rc := http.NewResponseController(w)
		rc.SetWriteDeadline(time.Now().Add(60 * time.Second))

		pyroscopeOnce.Do(func() {
			pyroscope.Start(pyroscope.Config{
				ApplicationName: "go-server-starter",
				ServerAddress:   utilities.GetEnvOrDefault("PYROSCOPE_ADDRESS", "http://localhost:4040"),
			})
		})
		pyroscope_pprof.Profile(w, r)
	})

	return mux
}
