// Package flagcast is the Go SDK: a thin client over the Evaluator gRPC service.
//
//	c, _ := flagcast.Dial("localhost:8202")
//	defer c.Close()
//	if c.BoolValue(ctx, "new-checkout", flagcast.Context{Key: userID}, false) { ... }
package flagcast

import (
	"context"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	flagcastv1 "github.com/mralaminahamed/flagcast/packages/shared/genproto/flagcast/v1"
)

// Context identifies the subject of an evaluation. Key drives percentage
// bucketing (e.g. a user id).
type Context struct {
	Key        string
	Attributes map[string]string
}

func (c Context) proto() *flagcastv1.Context {
	return &flagcastv1.Context{Key: c.Key, Attributes: c.Attributes}
}

type Client struct {
	conn *grpc.ClientConn
	c    flagcastv1.EvaluatorClient
}

// Dial connects to the evaluator. By default the connection is insecure
// (plaintext) — pass grpc dial options to add TLS. It does not block; the first
// RPC establishes the connection.
func Dial(addr string, opts ...grpc.DialOption) (*Client, error) {
	if len(opts) == 0 {
		opts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}
	// Propagate trace context to the evaluator on every call.
	opts = append(opts, grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	// Send the API key (from FLAGCAST_API_KEY) as x-api-key metadata when set.
	if key := os.Getenv("FLAGCAST_API_KEY"); key != "" {
		opts = append(opts,
			grpc.WithChainUnaryInterceptor(apiKeyUnary(key)),
			grpc.WithChainStreamInterceptor(apiKeyStream(key)),
		)
	}
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, c: flagcastv1.NewEvaluatorClient(conn)}, nil
}

func apiKeyUnary(key string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(metadata.AppendToOutgoingContext(ctx, "x-api-key", key), method, req, reply, cc, opts...)
	}
}

func apiKeyStream(key string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return streamer(metadata.AppendToOutgoingContext(ctx, "x-api-key", key), desc, cc, method, opts...)
	}
}

// Evaluate returns the decision and its reason, surfacing transport errors.
func (c *Client) Evaluate(ctx context.Context, flagKey string, ec Context) (bool, string, error) {
	resp, err := c.c.Evaluate(ctx, &flagcastv1.EvaluateRequest{FlagKey: flagKey, Context: ec.proto()})
	if err != nil {
		return false, "", err
	}
	return resp.GetValue(), resp.GetReason(), nil
}

// BoolValue returns the flag decision, falling back to def on any error or when
// the flag does not exist — the ergonomic call site for application code.
func (c *Client) BoolValue(ctx context.Context, flagKey string, ec Context, def bool) bool {
	v, reason, err := c.Evaluate(ctx, flagKey, ec)
	if err != nil || reason == "not_found" {
		return def
	}
	return v
}

// Change is a streamed flag update delivered to a Watch callback.
type Change struct {
	FlagKey string
	Value   bool
	Reason  string
	Deleted bool
}

// Watch streams flag changes for the context, invoking fn for the initial
// snapshot and every subsequent change. It blocks until ctx is cancelled or the
// stream errors, so callers typically run it in a goroutine.
func (c *Client) Watch(ctx context.Context, ec Context, fn func(Change)) error {
	stream, err := c.c.Watch(ctx, &flagcastv1.WatchRequest{Context: ec.proto()})
	if err != nil {
		return err
	}
	for {
		fc, err := stream.Recv()
		if err != nil {
			return err // io.EOF / context cancellation ends the watch
		}
		fn(Change{FlagKey: fc.GetFlagKey(), Value: fc.GetValue(), Reason: fc.GetReason(), Deleted: fc.GetDeleted()})
	}
}

// AllValues evaluates every flag for the context.
func (c *Client) AllValues(ctx context.Context, ec Context) (map[string]bool, error) {
	resp, err := c.c.EvaluateAll(ctx, &flagcastv1.EvaluateAllRequest{Context: ec.proto()})
	if err != nil {
		return nil, err
	}
	return resp.GetValues(), nil
}

func (c *Client) Close() error { return c.conn.Close() }
