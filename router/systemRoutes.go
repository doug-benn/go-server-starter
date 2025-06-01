package router

import (
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"runtime/debug"
	"time"

	"github.com/grafana/pyroscope-go"
	pyroscope_pprof "github.com/grafana/pyroscope-go/http/pprof"
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

// HandleGetDebug is a debug/profiling route for pyroscope
func HandleGetDebug() http.Handler {
	mux := http.NewServeMux()

	// Starting pyroscope profiler
	pyroscope.Start(pyroscope.Config{
		ApplicationName: "go-server-starter",
		ServerAddress:   "http://localhost:4040",
	})

	// Standard pprof routes (copied from /net/http/pprof)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// This route is special: note that we're using Pyroscope handler here
	mux.HandleFunc("/debug/pprof/profile", pyroscope_pprof.Profile)

	return mux
}
