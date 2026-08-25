// Command gateway is the flagcast control-plane entrypoint.
//
// Phase 0: a health/metrics stub on $PORT (default 8080) so the stack runs end
// to end. REST admin API (flags/segments), the gRPC evaluator, streaming, and
// the React console land in later phases. See PLAN.md.
package main

import (
	"github.com/mralaminahamed/flagcast/packages/shared/config"
	"github.com/mralaminahamed/flagcast/packages/shared/health"
	"github.com/mralaminahamed/flagcast/packages/shared/logger"
)

func main() {
	logger.InitLogger(logger.LoggerOptions{Level: config.Env("LOG_LEVEL", "info")})

	addr := health.AddrFromEnv(":8080")
	logger.Log.Info().Str("addr", addr).Msg("flagcast gateway listening (phase 0 stub)")
	if err := health.Serve("gateway", addr); err != nil {
		logger.Log.Fatal().Err(err).Msg("gateway stopped")
	}
}
