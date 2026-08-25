// Package tracing wires OpenTelemetry tracing for every service. It exports over
// OTLP/gRPC to OTEL_EXPORTER_OTLP_ENDPOINT; when that is unset, tracing is inert
// (no exporter, no sampling) so services run fine without a collector.
package tracing

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Init sets up the global tracer provider and W3C propagator for service. The
// returned shutdown func flushes spans; it is safe to call even when tracing is
// disabled.
func Init(ctx context.Context, service string) (func(context.Context) error, error) {
	// Always install the W3C propagator so trace context crosses HTTP/gRPC/NATS
	// even if this process isn't exporting.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		// No collector configured — leave the default no-op tracer in place.
		return func(context.Context) error { return nil }, nil
	}

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	exp, err := otlptracegrpc.New(cctx,
		otlptracegrpc.WithEndpoint(stripScheme(endpoint)),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	res, _ := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(service)))
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}

// stripScheme drops a leading http:// or https:// — otlptracegrpc wants host:port.
func stripScheme(s string) string {
	for _, p := range []string{"http://", "https://"} {
		if len(s) > len(p) && s[:len(p)] == p {
			return s[len(p):]
		}
	}
	return s
}
