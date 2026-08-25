// Package logger is a thin zerolog wrapper shared by every service.
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Log is the process-wide logger. InitLogger configures it at startup.
var Log = zerolog.New(os.Stdout).With().Timestamp().Logger()

type LoggerOptions struct {
	Level string // debug|info|warn|error (default info)
}

func InitLogger(opts LoggerOptions) {
	level, err := zerolog.ParseLevel(opts.Level)
	if err != nil || opts.Level == "" {
		level = zerolog.InfoLevel
	}
	zerolog.TimeFieldFormat = time.RFC3339
	Log = zerolog.New(os.Stdout).Level(level).With().Timestamp().Logger()
}
