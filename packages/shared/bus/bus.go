// Package bus is flagcast's lightweight NATS layer for flag-change events.
//
// Changes are ephemeral notifications: a subscriber that misses one re-syncs
// from Mongo at startup and on the next change, so core NATS pub/sub (not
// JetStream) is the right fit — fast fan-out, no durability overhead.
package bus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
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

func Connect(url string) (*Bus, error) {
	if url == "" {
		url = nats.DefaultURL
	}
	nc, err := nats.Connect(url, nats.MaxReconnects(-1), nats.ReconnectWait(time.Second))
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	return &Bus{nc: nc}, nil
}

// PublishJSON encodes v and publishes it to subject.
func (b *Bus) PublishJSON(subject string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return b.nc.Publish(subject, data)
}

// Subscribe registers handler for subject and returns an unsubscribe func.
func (b *Bus) Subscribe(subject string, handler func([]byte)) (func(), error) {
	sub, err := b.nc.Subscribe(subject, func(m *nats.Msg) { handler(m.Data) })
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
