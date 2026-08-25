// Package analyzer turns a flag's config, change history, and (optional)
// experiment metrics into a rollout recommendation. Deterministic stats always
// run; Claude adds the written reasoning when an API key is configured.
package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/mralaminahamed/flagcast/packages/shared/metrics"
	"github.com/mralaminahamed/flagcast/packages/shared/models"
	"github.com/mralaminahamed/flagcast/packages/shared/stats"
)

// Metrics is an optional A/B result to fold into the analysis.
type Metrics struct {
	Control   stats.Variant `json:"control"`
	Treatment stats.Variant `json:"treatment"`
}

// Request is one analysis input.
type Request struct {
	Flag    models.Flag         `json:"flag"`
	Recent  []models.AuditEntry `json:"recent"`
	Metrics *Metrics            `json:"metrics,omitempty"`
}

// Recommendation is the analysis output.
type Recommendation struct {
	Verdict string        `json:"verdict"` // ship | hold | iterate
	Summary string        `json:"summary"`
	Risks   []string      `json:"risks"`
	Stats   *stats.Result `json:"stats,omitempty"`
	Model   string        `json:"model,omitempty"` // "" when rule-based
}

type Analyzer struct {
	client *anthropic.Client // nil => rule-based only
	model  string
}

// New returns an Analyzer. With no apiKey it runs rule-based (no Claude).
func New(apiKey, model string) *Analyzer {
	if model == "" {
		model = "claude-opus-5"
	}
	a := &Analyzer{model: model}
	if apiKey != "" {
		c := anthropic.NewClient(option.WithAPIKey(apiKey))
		a.client = &c
	}
	return a
}

func (a *Analyzer) Analyze(ctx context.Context, req Request) (Recommendation, error) {
	var st *stats.Result
	if req.Metrics != nil {
		r := stats.TwoProportion(req.Metrics.Control, req.Metrics.Treatment)
		st = &r
	}
	rec := ruleBased(req, st)
	rec.Stats = st

	if a.client != nil {
		// Claude refines verdict/summary/risks; on any failure keep the rule-based rec.
		if refined, err := a.askClaude(ctx, req, st); err == nil {
			refined.Stats = st
			refined.Model = a.model
			rec = refined
		}
	}

	mode := "rule"
	if rec.Model != "" {
		mode = "claude"
	}
	metrics.Analyses.WithLabelValues(rec.Verdict, mode).Inc()
	return rec, nil
}

// ruleBased is the deterministic recommendation, and the fallback when Claude
// is unavailable.
func ruleBased(req Request, st *stats.Result) Recommendation {
	var risks []string
	if changes := len(req.Recent); changes >= 5 {
		risks = append(risks, fmt.Sprintf("%d recent changes — flag is churning", changes))
	}
	if st != nil {
		switch {
		case st.Significant && st.AbsoluteLift > 0:
			return Recommendation{Verdict: "ship", Risks: risks,
				Summary: fmt.Sprintf("Treatment lifts conversion %.1f%% (relative %.0f%%), significant at p=%.3g.",
					st.AbsoluteLift*100, st.RelativeLift*100, st.PValue)}
		case st.Significant && st.AbsoluteLift < 0:
			return Recommendation{Verdict: "hold", Risks: risks,
				Summary: fmt.Sprintf("Treatment lowers conversion %.1f%%, significant at p=%.3g. Do not ship.",
					-st.AbsoluteLift*100, st.PValue)}
		default:
			return Recommendation{Verdict: "iterate", Risks: risks,
				Summary: fmt.Sprintf("No significant difference yet (p=%.3g). Keep collecting or adjust the treatment.", st.PValue)}
		}
	}
	summary := "No experiment metrics supplied — review is based on config and change history."
	verdict := "iterate"
	if req.Flag.Enabled && req.Flag.Rollout >= 100 {
		verdict, summary = "ship", "Flag is fully rolled out and enabled."
	}
	return Recommendation{Verdict: verdict, Summary: summary, Risks: risks}
}

func (a *Analyzer) askClaude(ctx context.Context, req Request, st *stats.Result) (Recommendation, error) {
	ctxJSON, _ := json.MarshalIndent(struct {
		Flag    models.Flag         `json:"flag"`
		Recent  []models.AuditEntry `json:"recent_changes"`
		Metrics *Metrics            `json:"metrics,omitempty"`
		Stats   *stats.Result       `json:"stats,omitempty"`
	}{req.Flag, req.Recent, req.Metrics, st}, "", "  ")

	system := "You are a feature-flag rollout advisor. Given a flag's config, change history, and any " +
		"experiment stats, recommend whether to ship, hold, or iterate. Respond with ONLY a JSON object: " +
		`{"verdict":"ship|hold|iterate","summary":"one or two sentences","risks":["..."]}. No prose outside the JSON.`

	ctx, span := otel.Tracer("ai").Start(ctx, "claude.analyze")
	span.SetAttributes(attribute.String("model", a.model))
	defer span.End()

	adaptive := anthropic.ThinkingConfigAdaptiveParam{}
	resp, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(a.model),
		MaxTokens: 1024,
		System:    []anthropic.TextBlockParam{{Text: system}},
		Thinking:  anthropic.ThinkingConfigParamUnion{OfAdaptive: &adaptive},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(string(ctxJSON))),
		},
	})
	if err != nil {
		span.RecordError(err)
		return Recommendation{}, err
	}

	var text strings.Builder
	for _, block := range resp.Content {
		if b, ok := block.AsAny().(anthropic.TextBlock); ok {
			text.WriteString(b.Text)
		}
	}
	var rec Recommendation
	if err := json.Unmarshal([]byte(extractJSON(text.String())), &rec); err != nil {
		return Recommendation{}, err
	}
	if rec.Verdict == "" {
		return Recommendation{}, fmt.Errorf("empty verdict from model")
	}
	return rec, nil
}

// extractJSON pulls the first {...} object out of a response, tolerating any
// stray text around it.
func extractJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}
