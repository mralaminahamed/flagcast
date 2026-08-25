// Package grpcauth guards the evaluator's gRPC surface with a shared API key
// carried in the x-api-key metadata header. Mirrors the gateway's fail-closed
// policy: with no key set the server refuses all calls unless allowOpen (dev).
package grpcauth

import (
	"context"
	"crypto/subtle"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const headerKey = "x-api-key"

type Auth struct {
	key       string
	allowOpen bool
}

func New(key string, allowOpen bool) *Auth { return &Auth{key: key, allowOpen: allowOpen} }

func (a *Auth) authorize(ctx context.Context) error {
	if a.key == "" {
		if a.allowOpen {
			return nil
		}
		return status.Error(codes.Unauthenticated, "evaluator API key not configured")
	}
	md, _ := metadata.FromIncomingContext(ctx)
	vals := md.Get(headerKey)
	if len(vals) == 0 || subtle.ConstantTimeCompare([]byte(vals[0]), []byte(a.key)) != 1 {
		return status.Error(codes.Unauthenticated, "invalid api key")
	}
	return nil
}

func (a *Auth) Unary(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if err := a.authorize(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (a *Auth) Stream(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := a.authorize(ss.Context()); err != nil {
		return err
	}
	return handler(srv, ss)
}
