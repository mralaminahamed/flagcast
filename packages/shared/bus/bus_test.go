package bus

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// The gateway injects trace context into a message; the evaluator must extract
// the same trace so the two hops share one trace.
func TestTracePropagationRoundTrip(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})

	traceID, _ := trace.TraceIDFromHex("0123456789abcdef0123456789abcdef")
	spanID, _ := trace.SpanIDFromHex("0123456789abcdef")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled, Remote: true,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	h := nats.Header{}
	otel.GetTextMapPropagator().Inject(ctx, natsCarrier{h})
	if h.Get("traceparent") == "" {
		t.Fatal("no traceparent injected into headers")
	}

	got := trace.SpanContextFromContext(
		otel.GetTextMapPropagator().Extract(context.Background(), natsCarrier{h}),
	)
	if got.TraceID() != traceID {
		t.Errorf("extracted trace id = %s, want %s", got.TraceID(), traceID)
	}
	if !got.IsRemote() {
		t.Error("extracted span context should be remote")
	}
}
