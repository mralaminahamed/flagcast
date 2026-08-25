// Command evaluator serves flag decisions to SDK clients over gRPC. Flags come
// from MongoDB, fronted by a Redis cache kept fresh via NATS flag-change events.
// A separate HTTP port serves health/readiness/metrics.
package main

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/mralaminahamed/flagcast/apps/evaluator/internal/evalsvc"
	"github.com/mralaminahamed/flagcast/apps/evaluator/internal/flags"
	"github.com/mralaminahamed/flagcast/apps/evaluator/internal/grpcauth"
	"github.com/mralaminahamed/flagcast/apps/evaluator/internal/watch"
	"github.com/mralaminahamed/flagcast/packages/shared/bus"
	"github.com/mralaminahamed/flagcast/packages/shared/cache"
	"github.com/mralaminahamed/flagcast/packages/shared/config"
	flagcastv1 "github.com/mralaminahamed/flagcast/packages/shared/genproto/flagcast/v1"
	"github.com/mralaminahamed/flagcast/packages/shared/health"
	"github.com/mralaminahamed/flagcast/packages/shared/logger"
	"github.com/mralaminahamed/flagcast/packages/shared/metrics"
	"github.com/mralaminahamed/flagcast/packages/shared/store"
	"github.com/mralaminahamed/flagcast/packages/shared/tracing"
)

func main() {
	logger.InitLogger(logger.LoggerOptions{Level: config.Env("LOG_LEVEL", "info")})
	ctx := context.Background()

	if shutdown, err := tracing.Init(ctx, "flagcast-evaluator"); err != nil {
		logger.Log.Warn().Err(err).Msg("evaluator: tracing init failed")
	} else {
		defer shutdown(context.Background())
	}

	st, err := store.NewFlagStore(ctx, config.Env("MONGO_URI", "mongodb://localhost:27017"), config.Env("MONGO_DB", "flagcast"))
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("evaluator: cannot connect to MongoDB")
	}
	defer st.Close(ctx)

	checks := []health.Check{{Name: "mongo", Ping: st.Ping}}

	var fc *cache.FlagCache
	if url := config.Env("REDIS_URL", ""); url != "" {
		if c, err := cache.NewFlagCache(ctx, url); err != nil {
			logger.Log.Warn().Err(err).Msg("evaluator: Redis unavailable, reading flags from Mongo")
		} else {
			fc = c
			defer fc.Close()
			checks = append(checks, health.Check{Name: "redis", Ping: fc.Ping})
		}
	}

	repo := flags.New(st, fc)
	hub := watch.NewHub(repo)

	warm := func() {
		wctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := repo.Warm(wctx); err != nil {
			logger.Log.Warn().Err(err).Msg("evaluator: cache warm failed")
		}
	}

	if url := config.Env("NATS_URL", ""); url != "" {
		// Re-warm on reconnect: change events published while NATS was unreachable
		// are lost, so re-snapshot from Mongo once the connection returns.
		if b, err := bus.Connect(url, warm); err != nil {
			logger.Log.Warn().Err(err).Msg("evaluator: NATS unavailable, cache won't auto-refresh")
		} else {
			defer b.Close()
			checks = append(checks, health.Check{Name: "nats", Ping: b.Ping})
			// Subscribe BEFORE the initial warm so no change is missed in between.
			_, err := b.Subscribe(bus.SubjectFlagChanged, func(mctx context.Context, data []byte) {
				var evt bus.FlagChanged
				if json.Unmarshal(data, &evt) != nil {
					return
				}
				metrics.ChangesProcessed.Inc()
				// mctx carries the trace extracted from the message headers, so the
				// refresh + fan-out are in the same trace as the gateway mutation.
				repo.OnChange(mctx, evt)
				hub.Broadcast(mctx, evt)
			})
			if err != nil {
				logger.Log.Warn().Err(err).Msg("evaluator: subscribe flag.changed")
			}
		}
	}

	// Initial snapshot, then a periodic full re-sync that deterministically
	// repairs any drift from a missed event (a dropped flag.changed never leaves
	// the cache permanently stale).
	warm()
	resync := time.Duration(config.EnvInt("RESYNC_SECONDS", 60)) * time.Second
	go func() {
		t := time.NewTicker(resync)
		defer t.Stop()
		for range t.C {
			warm()
		}
	}()

	grpcAddr := config.Env("GRPC_ADDR", ":50051")
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Log.Fatal().Err(err).Str("addr", grpcAddr).Msg("evaluator: listen")
	}
	// Closed on shutdown so open Watch streams end and GracefulStop can complete.
	shutdownCh := make(chan struct{})
	allowOpen := os.Getenv("ALLOW_OPEN_API") == "true"
	auth := grpcauth.New(os.Getenv("EVALUATOR_API_KEY"), allowOpen)
	if os.Getenv("EVALUATOR_API_KEY") == "" && !allowOpen {
		logger.Log.Error().Msg("evaluator: EVALUATOR_API_KEY unset and ALLOW_OPEN_API!=true — gRPC refuses all calls")
	}
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(auth.Unary),
		grpc.ChainStreamInterceptor(auth.Stream),
	)
	flagcastv1.RegisterEvaluatorServer(srv, evalsvc.New(repo, hub, shutdownCh))
	if allowOpen {
		reflection.Register(srv) // introspection is a dev convenience; off in prod
	}

	go func() {
		addr := health.AddrFromEnv(":8081")
		logger.Log.Info().Str("addr", addr).Msg("evaluator health listening")
		_ = health.Serve("evaluator", addr, checks...)
	}()
	go func() {
		logger.Log.Info().Str("addr", grpcAddr).Msg("evaluator gRPC listening")
		if err := srv.Serve(lis); err != nil {
			logger.Log.Error().Err(err).Msg("evaluator: grpc serve")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Log.Info().Msg("evaluator shutting down")
	close(shutdownCh) // end open Watch streams
	// Time-box the graceful stop; fall back to a hard Stop so a wedged stream
	// can't block shutdown indefinitely.
	done := make(chan struct{})
	go func() { srv.GracefulStop(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		logger.Log.Warn().Msg("evaluator: graceful stop timed out, forcing")
		srv.Stop()
	}
}
