// Package config provides small env helpers shared across services.
package config

import (
	"os"
	"strconv"
)

// Env returns the value of key, or fallback when unset/empty.
func Env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// EnvInt parses key as an int, or returns fallback.
func EnvInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return fallback
}

// EnvBool reports whether key is set to "true".
func EnvBool(key string) bool {
	return os.Getenv(key) == "true"
}
