package eval

import (
	"fmt"
	"testing"

	"github.com/mralaminahamed/flagcast/packages/shared/models"
)

func TestEvaluateGates(t *testing.T) {
	cases := []struct {
		name       string
		flag       models.Flag
		wantValue  bool
		wantReason string
	}{
		{"disabled", models.Flag{Key: "f", Enabled: false, Rollout: 100}, false, "disabled"},
		{"full", models.Flag{Key: "f", Enabled: true, Rollout: 100}, true, "enabled"},
		{"zero", models.Flag{Key: "f", Enabled: true, Rollout: 0}, false, "rollout"},
	}
	for _, c := range cases {
		v, r := Evaluate(c.flag, "user-1")
		if v != c.wantValue || r != c.wantReason {
			t.Errorf("%s: got (%v,%q) want (%v,%q)", c.name, v, r, c.wantValue, c.wantReason)
		}
	}
}

func TestEvaluateDeterministic(t *testing.T) {
	f := models.Flag{Key: "checkout", Enabled: true, Rollout: 50}
	first, _ := Evaluate(f, "user-42")
	for i := 0; i < 100; i++ {
		if v, _ := Evaluate(f, "user-42"); v != first {
			t.Fatalf("non-deterministic result for same context")
		}
	}
}

func TestEvaluateRolloutDistribution(t *testing.T) {
	f := models.Flag{Key: "checkout", Enabled: true, Rollout: 50}
	on := 0
	const n = 5000
	for i := 0; i < n; i++ {
		if v, _ := Evaluate(f, fmt.Sprintf("user-%d", i)); v {
			on++
		}
	}
	pct := float64(on) / n * 100
	if pct < 42 || pct > 58 {
		t.Errorf("50%% rollout produced %.1f%% on (expected ~50)", pct)
	}
}
