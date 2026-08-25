// Package bus is flagcast's lightweight NATS layer for flag-change events.
//
// Changes are ephemeral notifications: a subscriber that misses one re-syncs
// from Mongo at startup and on the next change, so core NATS pub/sub (not
// JetStream) is the right fit — fast fan-out, no durability overhead. W3C trace
// context rides in message headers so a trace spans the gateway → evaluator hop.
package bus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
)

// SubjectFlagChanged carries flag mutations from the gateway to evaluators.
const SubjectFlagChanged = "flag.changed"

// FlagChanged is published when a flag is created, updated, or deleted.
type FlagChanged struct {
	Key    string `json:"key"`
	Action string `json:"action"` // created | updated | deleted
}

type Bus struct {
	nc *nats.Conn
}

// Connect dials NATS. Any onReconnect callbacks fire after the client
// re-establishes a dropped connection — the caller uses this to re-sync state
// that may have changed (and whose change events were lost) during the outage.
func Connect(url string, onReconnect ...func()) (*Bus, error) {
	if url == "" {
		url = nats.DefaultURL
	}
	opts := []nats.Option{nats.MaxReconnects(-1), nats.ReconnectWait(time.Second)}
	if len(onReconnect) > 0 {
		opts = append(opts, nats.ReconnectHandler(func(*nats.Conn) {
			for _, fn := range onReconnect {
				fn()
			}
		}))
	}
	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	return &Bus{nc: nc}, nil
}

// natsCarrier adapts nats.Header to the OTel TextMapCarrier interface.
type natsCarrier struct{ h nats.Header }

func (c natsCarrier) Get(key string) string {
	if v := c.h[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}
func (c natsCarrier) Set(key, value string) { c.h[key] = []string{value} }
func (c natsCarrier) Keys() []string {
	ks := make([]string, 0, len(c.h))
	for k := range c.h {
		ks = append(ks, k)
	}
	return ks
}

// PublishJSON encodes v and publishes it to subject, injecting the trace context
// from ctx into the message headers.
func (b *Bus) PublishJSON(ctx context.Context, subject string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	msg := &nats.Msg{Subject: subject, Data: data, Header: nats.Header{}}
	otel.GetTextMapPropagator().Inject(ctx, natsCarrier{msg.Header})
	return b.nc.PublishMsg(msg)
}

// Subscribe registers handler for subject and returns an unsubscribe func. The
// handler receives a context carrying the extracted trace and a consumer span.
func (b *Bus) Subscribe(subject string, handler func(context.Context, []byte)) (func(), error) {
	sub, err := b.nc.Subscribe(subject, func(m *nats.Msg) {
		ctx := context.Background()
		if m.Header != nil {
			ctx = otel.GetTextMapPropagator().Extract(ctx, natsCarrier{m.Header})
		}
		ctx, span := otel.Tracer("bus").Start(ctx, subject+" consume")
		defer span.End()
		handler(ctx, m.Data)
	})
	if err != nil {
		return nil, err
	}
	return func() { _ = sub.Unsubscribe() }, nil
}

// Ping reports whether the connection is healthy.
func (b *Bus) Ping(context.Context) error {
	if b.nc == nil || !b.nc.IsConnected() {
		return fmt.Errorf("nats not connected")
	}
	return nil
}

func (b *Bus) Close() {
	if b.nc != nil {
		_ = b.nc.Drain()
	}
}
