// Package logger is a thin zerolog wrapper shared by every service.
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Log is the process-wide logger. InitLogger configures it at startup.
var Log = zerolog.New(os.Stdout).With().Timestamp().Logger()

type LoggerOptions struct {
	Level  string // debug|info|warn|error (default info)
	Stderr bool   // write to stderr instead of stdout (e.g. stdio MCP servers)
}

func InitLogger(opts LoggerOptions) {
	level, err := zerolog.ParseLevel(opts.Level)
	if err != nil || opts.Level == "" {
		level = zerolog.InfoLevel
	}
	zerolog.TimeFieldFormat = time.RFC3339
	var w io.Writer = os.Stdout
	if opts.Stderr {
		w = os.Stderr
	}
	Log = zerolog.New(w).Level(level).With().Timestamp().Logger()
}
