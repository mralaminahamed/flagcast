// Package stats computes A/B experiment statistics (deterministic, no LLM).
package stats

import "math"

// Variant is one arm of an experiment.
type Variant struct {
	Exposures   int `json:"exposures"`
	Conversions int `json:"conversions"`
}

// Result is a two-proportion comparison of treatment against control.
type Result struct {
	ControlRate   float64 `json:"control_rate"`
	TreatmentRate float64 `json:"treatment_rate"`
	AbsoluteLift  float64 `json:"absolute_lift"`
	RelativeLift  float64 `json:"relative_lift"`
	ZScore        float64 `json:"z_score"`
	PValue        float64 `json:"p_value"`
	Significant   bool    `json:"significant"` // p < 0.05, two-tailed
}

// TwoProportion runs a two-proportion z-test comparing treatment vs control.
// With no exposures it returns a zero Result (nothing to conclude).
func TwoProportion(control, treatment Variant) Result {
	if control.Exposures == 0 || treatment.Exposures == 0 {
		return Result{}
	}
	nc, nt := float64(control.Exposures), float64(treatment.Exposures)
	pc := float64(control.Conversions) / nc
	pt := float64(treatment.Conversions) / nt

	res := Result{
		ControlRate:   pc,
		TreatmentRate: pt,
		AbsoluteLift:  pt - pc,
	}
	if pc > 0 {
		res.RelativeLift = (pt - pc) / pc
	}

	pooled := float64(control.Conversions+treatment.Conversions) / (nc + nt)
	se := math.Sqrt(pooled * (1 - pooled) * (1/nc + 1/nt))
	if se > 0 {
		res.ZScore = (pt - pc) / se
		res.PValue = 2 * (1 - normCDF(math.Abs(res.ZScore)))
		res.Significant = res.PValue < 0.05
	}
	return res
}

// normCDF is the standard normal cumulative distribution function.
func normCDF(x float64) float64 {
	return 0.5 * math.Erfc(-x/math.Sqrt2)
}
