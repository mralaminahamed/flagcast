// Command evaluator serves flag decisions to SDK clients over gRPC, reading flag
// config from MongoDB. A separate HTTP port serves health/readiness/metrics.
package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/mralaminahamed/flagcast/apps/evaluator/internal/evalsvc"
	"github.com/mralaminahamed/flagcast/packages/shared/config"
	flagcastv1 "github.com/mralaminahamed/flagcast/packages/shared/genproto/flagcast/v1"
	"github.com/mralaminahamed/flagcast/packages/shared/health"
	"github.com/mralaminahamed/flagcast/packages/shared/logger"
	"github.com/mralaminahamed/flagcast/packages/shared/store"
)

func main() {
	logger.InitLogger(logger.LoggerOptions{Level: config.Env("LOG_LEVEL", "info")})

	st, err := store.NewFlagStore(context.Background(), config.Env("MONGO_URI", "mongodb://localhost:27017"), config.Env("MONGO_DB", "flagcast"))
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("evaluator: cannot connect to MongoDB")
	}
	defer st.Close(context.Background())

	grpcAddr := config.Env("GRPC_ADDR", ":50051")
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Log.Fatal().Err(err).Str("addr", grpcAddr).Msg("evaluator: listen")
	}
	srv := grpc.NewServer()
	flagcastv1.RegisterEvaluatorServer(srv, evalsvc.New(st))
	reflection.Register(srv) // lets grpcurl/tools introspect the API

	go func() {
		addr := health.AddrFromEnv(":8081")
		logger.Log.Info().Str("addr", addr).Msg("evaluator health listening")
		_ = health.Serve("evaluator", addr, health.Check{Name: "mongo", Ping: st.Ping})
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
	srv.GracefulStop()
}
