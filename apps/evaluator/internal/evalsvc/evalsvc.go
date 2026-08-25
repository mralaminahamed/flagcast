// Package evalsvc implements the Evaluator gRPC service.
package evalsvc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mralaminahamed/flagcast/packages/shared/eval"
	flagcastv1 "github.com/mralaminahamed/flagcast/packages/shared/genproto/flagcast/v1"
	"github.com/mralaminahamed/flagcast/packages/shared/models"
	"github.com/mralaminahamed/flagcast/packages/shared/store"
)

// FlagSource supplies flags for evaluation (Mongo, or the cache-fronted repo).
type FlagSource interface {
	Get(ctx context.Context, key string) (models.Flag, error)
	List(ctx context.Context) ([]models.Flag, error)
}

// Server answers flag-evaluation RPCs, reading flag config from a FlagSource.
type Server struct {
	flagcastv1.UnimplementedEvaluatorServer
	src FlagSource
}

func New(src FlagSource) *Server { return &Server{src: src} }

func (s *Server) Evaluate(ctx context.Context, req *flagcastv1.EvaluateRequest) (*flagcastv1.EvaluateResponse, error) {
	if req.GetFlagKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "flag_key is required")
	}
	f, err := s.src.Get(ctx, req.GetFlagKey())
	if err != nil {
		// An unknown flag is not an error to the caller — it evaluates to off.
		if errors.Is(err, store.ErrNotFound) {
			return &flagcastv1.EvaluateResponse{FlagKey: req.GetFlagKey(), Value: false, Reason: "not_found"}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	value, reason := eval.Evaluate(f, req.GetContext().GetKey())
	return &flagcastv1.EvaluateResponse{FlagKey: f.Key, Value: value, Reason: reason}, nil
}

func (s *Server) EvaluateAll(ctx context.Context, req *flagcastv1.EvaluateAllRequest) (*flagcastv1.EvaluateAllResponse, error) {
	flags, err := s.src.List(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	ctxKey := req.GetContext().GetKey()
	values := make(map[string]bool, len(flags))
	for _, f := range flags {
		v, _ := eval.Evaluate(f, ctxKey)
		values[f.Key] = v
	}
	return &flagcastv1.EvaluateAllResponse{Values: values}, nil
}
