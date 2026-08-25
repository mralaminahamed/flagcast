// Command ai serves flag rollout analysis over HTTP. It reads flag config and
// change history from Mongo and, when ANTHROPIC_API_KEY is set, uses Claude to
// write the recommendation; otherwise it falls back to deterministic rules.
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mralaminahamed/flagcast/apps/ai/internal/analyzer"
	"github.com/mralaminahamed/flagcast/packages/shared/config"
	"github.com/mralaminahamed/flagcast/packages/shared/health"
	"github.com/mralaminahamed/flagcast/packages/shared/logger"
	"github.com/mralaminahamed/flagcast/packages/shared/store"
)

func main() {
	logger.InitLogger(logger.LoggerOptions{Level: config.Env("LOG_LEVEL", "info")})
	ctx := context.Background()

	st, err := store.NewFlagStore(ctx, config.Env("MONGO_URI", "mongodb://localhost:27017"), config.Env("MONGO_DB", "flagcast"))
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("ai: cannot connect to MongoDB")
	}
	defer st.Close(ctx)

	az := analyzer.New(os.Getenv("ANTHROPIC_API_KEY"), config.Env("ANTHROPIC_MODEL", "claude-opus-5"))
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		logger.Log.Warn().Msg("ai: ANTHROPIC_API_KEY unset — using rule-based analysis")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", health.LivenessHandler("ai"))
	mux.HandleFunc("/ready", health.ReadyHandler("ai", health.Check{Name: "mongo", Ping: st.Ping}))
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/analyze", analyzeHandler(st, az))

	addr := health.AddrFromEnv(":8090")
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		logger.Log.Info().Str("addr", addr).Msg("ai listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error().Err(err).Msg("ai server stopped")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Log.Info().Msg("ai shutting down")
	sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(sctx)
}

func analyzeHandler(st *store.FlagStore, az *analyzer.Analyzer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		var body struct {
			FlagKey string            `json:"flag_key"`
			Metrics *analyzer.Metrics `json:"metrics"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if body.FlagKey == "" {
			writeErr(w, http.StatusBadRequest, "flag_key is required")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 40*time.Second)
		defer cancel()
		flag, err := st.Get(ctx, body.FlagKey)
		if err != nil {
			writeErr(w, http.StatusNotFound, "flag not found")
			return
		}
		recent, _ := st.ListAudit(ctx, body.FlagKey, 20)

		rec, err := az.Analyze(ctx, analyzer.Request{Flag: flag, Recent: recent, Metrics: body.Metrics})
		if err != nil {
			writeErr(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, rec)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
