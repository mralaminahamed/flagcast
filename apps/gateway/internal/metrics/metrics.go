// Package metrics exposes Prometheus HTTP metrics for the gateway.
package metrics

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flagcast_http_requests_total",
		Help: "HTTP requests by method, route, and status.",
	}, []string{"method", "route", "status"})

	duration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "flagcast_http_request_duration_seconds",
		Help:    "HTTP request duration by method and route.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
)

// Middleware records request count and duration keyed by the matched route.
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			route := c.Path()
			if route == "" {
				route = "unmatched"
			}
			requests.WithLabelValues(c.Request().Method, route, strconv.Itoa(c.Response().Status)).Inc()
			duration.WithLabelValues(c.Request().Method, route).Observe(time.Since(start).Seconds())
			return err
		}
	}
}
