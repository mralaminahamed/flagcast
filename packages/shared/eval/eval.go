// Package eval is the flag decision logic, shared by the evaluator service (and
// reusable for future SDK-side local evaluation).
package eval

import (
	"hash/fnv"

	"github.com/mralaminahamed/flagcast/packages/shared/models"
)

// Evaluate decides a flag for a context key and returns the value plus a reason.
// A disabled flag is off; a 100% rollout is on; otherwise the context is placed
// in a stable 0-99 bucket and is on when its bucket is below the rollout percent.
func Evaluate(f models.Flag, contextKey string) (bool, string) {
	if !f.Enabled {
		return false, "disabled"
	}
	if f.Rollout >= 100 {
		return true, "enabled"
	}
	if f.Rollout <= 0 {
		return false, "rollout"
	}
	if bucket(f.Key, contextKey) < f.Rollout {
		return true, "rollout"
	}
	return false, "rollout"
}

// bucket maps (flagKey, contextKey) to a stable value in [0,100). Keying the
// hash by flag means the same context lands in different buckets per flag, so
// rollouts across flags are independent rather than correlated.
func bucket(flagKey, contextKey string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(flagKey))
	_, _ = h.Write([]byte{':'})
	_, _ = h.Write([]byte(contextKey))
	return int(h.Sum32() % 100)
}
