// Package health serves liveness (/health), readiness (/ready), and /metrics.
// Liveness is always ok; readiness runs the given dependency checks so an
// orchestrator does not route traffic to a pod that cannot reach its deps.
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Check is a named dependency probe.
type Check struct {
	Name string
	Ping func(context.Context) error
}

func Serve(service, addr string, checks ...Check) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", LivenessHandler(service))
	mux.HandleFunc("/ready", ReadyHandler(service, checks...))
	mux.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, mux)
}

// LivenessHandler always reports ok — the process is running.
func LivenessHandler(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": service})
	}
}

// ReadyHandler runs the dependency checks and returns 503 if any fail.
func ReadyHandler(service string, checks ...Check) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		deps := map[string]string{}
		ready := true
		for _, c := range checks {
			if err := c.Ping(ctx); err != nil {
				deps[c.Name] = err.Error()
				ready = false
			} else {
				deps[c.Name] = "ok"
			}
		}
		code, status := http.StatusOK, "ready"
		if !ready {
			code, status = http.StatusServiceUnavailable, "not ready"
		}
		writeJSON(w, code, map[string]any{"status": status, "service": service, "deps": deps})
	}
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

// AddrFromEnv returns ":$PORT" when PORT is set, otherwise fallback.
func AddrFromEnv(fallback string) string {
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return fallback
}
