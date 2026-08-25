package grpcauth

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ctxWithKey(k string) context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-api-key", k))
}

func code(err error) codes.Code { return status.Code(err) }

func TestAuthorize(t *testing.T) {
	cases := []struct {
		name      string
		key       string
		allowOpen bool
		ctx       context.Context
		want      codes.Code
	}{
		{"fail-closed: no key, no open", "", false, context.Background(), codes.Unauthenticated},
		{"open: no key, allowOpen", "", true, context.Background(), codes.OK},
		{"keyed: missing header", "k", false, context.Background(), codes.Unauthenticated},
		{"keyed: wrong key", "k", false, ctxWithKey("nope"), codes.Unauthenticated},
		{"keyed: right key", "k", false, ctxWithKey("k"), codes.OK},
	}
	for _, c := range cases {
		got := code(New(c.key, c.allowOpen).authorize(c.ctx))
		if got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
