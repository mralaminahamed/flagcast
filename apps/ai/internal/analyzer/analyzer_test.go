package analyzer

import (
	"context"
	"testing"

	"github.com/mralaminahamed/flagcast/packages/shared/models"
	"github.com/mralaminahamed/flagcast/packages/shared/stats"
)

func TestAnalyzeRuleBased(t *testing.T) {
	a := New("", "") // no key → rule-based
	ctx := context.Background()

	// Significant positive lift → ship.
	rec, _ := a.Analyze(ctx, Request{
		Flag:    models.Flag{Key: "f", Enabled: true, Rollout: 50},
		Metrics: &Metrics{Control: stats.Variant{Exposures: 1000, Conversions: 100}, Treatment: stats.Variant{Exposures: 1000, Conversions: 160}},
	})
	if rec.Verdict != "ship" || rec.Model != "" || rec.Stats == nil || !rec.Stats.Significant {
		t.Fatalf("significant lift: got %+v", rec)
	}

	// Significant negative lift → hold.
	rec, _ = a.Analyze(ctx, Request{
		Flag:    models.Flag{Key: "f", Enabled: true, Rollout: 50},
		Metrics: &Metrics{Control: stats.Variant{Exposures: 1000, Conversions: 160}, Treatment: stats.Variant{Exposures: 1000, Conversions: 100}},
	})
	if rec.Verdict != "hold" {
		t.Fatalf("negative lift: got %q", rec.Verdict)
	}

	// No metrics, fully rolled out → ship; churn risk surfaced.
	rec, _ = a.Analyze(ctx, Request{
		Flag:   models.Flag{Key: "f", Enabled: true, Rollout: 100},
		Recent: make([]models.AuditEntry, 6),
	})
	if rec.Verdict != "ship" || len(rec.Risks) == 0 {
		t.Fatalf("no-metrics: got %+v", rec)
	}
}

func TestExtractJSON(t *testing.T) {
	cases := map[string]string{
		`{"a":1}`:                 `{"a":1}`,
		"prefix {\"a\":1} suffix": `{"a":1}`,
		"```json\n{\"a\":1}\n```": `{"a":1}`,
		"no json here":            "no json here",
	}
	for in, want := range cases {
		if got := extractJSON(in); got != want {
			t.Errorf("extractJSON(%q)=%q want %q", in, got, want)
		}
	}
}
