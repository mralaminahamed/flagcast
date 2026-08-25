// Command gateway is the flagcast control-plane: the REST admin API for flags
// and the change audit log, backed by MongoDB.
package main

import (
	"context"

	"github.com/mralaminahamed/flagcast/apps/gateway/internal/server"
	"github.com/mralaminahamed/flagcast/packages/shared/config"
	"github.com/mralaminahamed/flagcast/packages/shared/health"
	"github.com/mralaminahamed/flagcast/packages/shared/logger"
	"github.com/mralaminahamed/flagcast/packages/shared/store"
)

func main() {
	logger.InitLogger(logger.LoggerOptions{Level: config.Env("LOG_LEVEL", "info")})

	uri := config.Env("MONGO_URI", "mongodb://localhost:27017")
	st, err := store.NewFlagStore(context.Background(), uri, config.Env("MONGO_DB", "flagcast"))
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("gateway: cannot connect to MongoDB")
	}
	defer st.Close(context.Background())

	addr := health.AddrFromEnv(":8080")
	if err := server.Run(st, addr); err != nil {
		logger.Log.Fatal().Err(err).Msg("gateway: run")
	}
}
