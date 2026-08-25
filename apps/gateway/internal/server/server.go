// Package server builds and runs the gateway's Echo HTTP server.
package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"

	"github.com/mralaminahamed/flagcast/apps/gateway/internal/handler"
	"github.com/mralaminahamed/flagcast/apps/gateway/internal/metrics"
	"github.com/mralaminahamed/flagcast/packages/shared/logger"
	"github.com/mralaminahamed/flagcast/packages/shared/store"
)

const rateLimitPerSec = 20

// Publisher is re-exported so callers can wire a bus without importing handler.
type Publisher = handler.Publisher

// New builds a configured Echo instance with middleware and routes. pub may be
// nil (flag-change events are then skipped).
func New(s *store.FlagStore, pub handler.Publisher) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	// Log the path, not the full URI, so query params (incl. any api key) stay
	// out of access logs.
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339}","id":"${id}","method":"${method}","path":"${path}",` +
			`"status":${status},"latency":"${latency_human}","error":"${error}"}` + "\n",
	}))
	e.Use(middleware.BodyLimit("1M"))
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(rateLimitPerSec))))
	e.Use(metrics.Middleware())
	if origins := os.Getenv("CORS_ORIGINS"); origins != "" {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: strings.Split(origins, ",")}))
	} else {
		logger.Log.Warn().Msg("gateway: CORS_ORIGINS unset — allowing all origins (dev)")
		e.Use(middleware.CORS())
	}

	h := handler.New(s, pub)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"status": "ok", "service": "gateway"})
	})
	e.GET("/ready", func(c echo.Context) error {
		if err := s.Ping(c.Request().Context()); err != nil {
			return c.JSON(http.StatusServiceUnavailable, echo.Map{"status": "not ready", "deps": echo.Map{"mongo": err.Error()}})
		}
		return c.JSON(http.StatusOK, echo.Map{"status": "ready", "deps": echo.Map{"mongo": "ok"}})
	})
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	api := e.Group("/api")
	if key := os.Getenv("GATEWAY_API_KEY"); key != "" {
		api.Use(apiKey(key))
	} else {
		logger.Log.Warn().Msg("gateway: GATEWAY_API_KEY unset — /api is unauthenticated")
	}
	api.GET("/flags", h.List)
	api.POST("/flags", h.Create)
	api.GET("/flags/:key", h.Get)
	api.PUT("/flags/:key", h.Update)
	api.DELETE("/flags/:key", h.Delete)
	api.GET("/audit", h.Audit)

	return e
}

// Run starts the server and blocks until SIGINT/SIGTERM, then shuts down.
func Run(s *store.FlagStore, pub handler.Publisher, addr string) error {
	e := New(s, pub)
	e.Server.ReadHeaderTimeout = 10 * time.Second
	e.Server.ReadTimeout = 30 * time.Second
	e.Server.IdleTimeout = 120 * time.Second

	go func() {
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error().Err(err).Msg("gateway server stopped")
		}
	}()
	logger.Log.Info().Str("addr", addr).Msg("gateway listening")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Log.Info().Msg("gateway shutting down")
	return e.Shutdown(ctx)
}
