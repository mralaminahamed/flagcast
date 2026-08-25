// Package validation holds shared input validation.
package validation

import (
	"fmt"
	"regexp"
)

var keyRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-_.]{0,63}$`)

// FlagKey checks a flag key is a safe, stable identifier.
func FlagKey(key string) error {
	if !keyRe.MatchString(key) {
		return fmt.Errorf("invalid key: must match %s", keyRe.String())
	}
	return nil
}

// Rollout checks the percentage is within 0-100.
func Rollout(pct int) error {
	if pct < 0 || pct > 100 {
		return fmt.Errorf("rollout must be between 0 and 100")
	}
	return nil
}
