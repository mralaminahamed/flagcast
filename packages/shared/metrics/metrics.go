// Package metrics defines flagcast's business metrics on the default Prometheus
// registry. Every service exposes them via health.Serve's /metrics.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Evaluations counts flag decisions by reason (enabled/disabled/rollout/not_found).
	Evaluations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flagcast_evaluations_total",
		Help: "Flag evaluations by reason.",
	}, []string{"reason"})

	// CacheOps counts flag-cache reads by result (hit/miss).
	CacheOps = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flagcast_flag_cache_ops_total",
		Help: "Flag cache reads by result.",
	}, []string{"result"})

	// ChangesProcessed counts flag.changed events the evaluator applied.
	ChangesProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flagcast_flag_changes_processed_total",
		Help: "flag.changed events processed by the evaluator.",
	})

	// WatchStreams tracks currently-connected Watch streams.
	WatchStreams = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "flagcast_watch_streams",
		Help: "Currently connected Watch streams.",
	})

	// WatchDrops counts updates dropped to slow watchers.
	WatchDrops = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flagcast_watch_drops_total",
		Help: "Watch updates dropped because a client was too slow.",
	})

	// Analyses counts AI analyses by verdict and mode (claude/rule).
	Analyses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flagcast_analyses_total",
		Help: "Rollout analyses by verdict and mode.",
	}, []string{"verdict", "mode"})
)
