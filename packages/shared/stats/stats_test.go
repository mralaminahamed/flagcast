package stats

import (
	"math"
	"testing"
)

func approx(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

func TestTwoProportionSignificant(t *testing.T) {
	// Strong, clearly significant lift.
	r := TwoProportion(Variant{Exposures: 1000, Conversions: 100}, Variant{Exposures: 1000, Conversions: 160})
	if !approx(r.ControlRate, 0.10, 1e-9) || !approx(r.TreatmentRate, 0.16, 1e-9) {
		t.Fatalf("rates: %+v", r)
	}
	if !approx(r.AbsoluteLift, 0.06, 1e-9) || !approx(r.RelativeLift, 0.6, 1e-9) {
		t.Fatalf("lift: %+v", r)
	}
	if !r.Significant || r.PValue > 0.001 {
		t.Fatalf("expected strong significance, got p=%.4g", r.PValue)
	}
}

func TestTwoProportionNotSignificant(t *testing.T) {
	// Tiny difference on small samples — not significant.
	r := TwoProportion(Variant{Exposures: 100, Conversions: 50}, Variant{Exposures: 100, Conversions: 52})
	if r.Significant {
		t.Fatalf("did not expect significance, p=%.4g", r.PValue)
	}
}

func TestTwoProportionEmpty(t *testing.T) {
	if r := TwoProportion(Variant{}, Variant{Exposures: 10, Conversions: 1}); r != (Result{}) {
		t.Fatalf("empty control should yield zero result, got %+v", r)
	}
}
