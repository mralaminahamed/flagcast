package validation

import "testing"

func TestFlagKey(t *testing.T) {
	valid := []string{"new-checkout", "feature.x", "a", "flag_1", "a-b.c_d"}
	invalid := []string{"", "-bad", "Bad", "with space", "way-" + string(make([]byte, 70))}
	for _, k := range valid {
		if err := FlagKey(k); err != nil {
			t.Errorf("FlagKey(%q) = %v, want nil", k, err)
		}
	}
	for _, k := range invalid {
		if err := FlagKey(k); err == nil {
			t.Errorf("FlagKey(%q) = nil, want error", k)
		}
	}
}

func TestRollout(t *testing.T) {
	for _, ok := range []int{0, 1, 50, 100} {
		if err := Rollout(ok); err != nil {
			t.Errorf("Rollout(%d) = %v, want nil", ok, err)
		}
	}
	for _, bad := range []int{-1, 101, 1000} {
		if err := Rollout(bad); err == nil {
			t.Errorf("Rollout(%d) = nil, want error", bad)
		}
	}
}
