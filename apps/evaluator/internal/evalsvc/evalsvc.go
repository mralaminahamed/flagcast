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

// Watchers registers a Watch stream and returns its change channel, an
// unregister func, and ok=false when the watcher cap is reached.
type Watchers interface {
	Register(contextKey string) (<-chan *flagcastv1.FlagChange, func(), bool)
}

// Server answers flag-evaluation RPCs, reading flag config from a FlagSource.
type Server struct {
	flagcastv1.UnimplementedEvaluatorServer
	src      FlagSource
	hub      Watchers        // may be nil (Watch unavailable)
	shutdown <-chan struct{} // closed on server shutdown to end Watch streams
}

func New(src FlagSource, hub Watchers, shutdown <-chan struct{}) *Server {
	return &Server{src: src, hub: hub, shutdown: shutdown}
}

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

// Watch streams a snapshot of all flag values for the request's context, then
// pushes a FlagChange whenever a flag changes, until the client disconnects.
func (s *Server) Watch(req *flagcastv1.WatchRequest, stream flagcastv1.Evaluator_WatchServer) error {
	if s.hub == nil {
		return status.Error(codes.Unavailable, "watch not enabled")
	}
	ctxKey := req.GetContext().GetKey()

	// Register before the snapshot so no change is missed between them.
	ch, unregister, ok := s.hub.Register(ctxKey)
	if !ok {
		return status.Error(codes.ResourceExhausted, "too many watchers")
	}
	defer unregister()

	flags, err := s.src.List(stream.Context())
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	for _, f := range flags {
		v, reason := eval.Evaluate(f, ctxKey)
		if err := stream.Send(&flagcastv1.FlagChange{FlagKey: f.Key, Value: v, Reason: reason}); err != nil {
			return err
		}
	}

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-s.shutdown:
			// Server is stopping — end the stream so GracefulStop can complete.
			return status.Error(codes.Unavailable, "server shutting down")
		case fc := <-ch:
			if err := stream.Send(fc); err != nil {
				return err
			}
		}
	}
}
