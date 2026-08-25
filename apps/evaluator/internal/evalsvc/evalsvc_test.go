package evalsvc

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	flagcastv1 "github.com/mralaminahamed/flagcast/packages/shared/genproto/flagcast/v1"
	"github.com/mralaminahamed/flagcast/packages/shared/models"
	"github.com/mralaminahamed/flagcast/packages/shared/store"
)

type fakeSrc struct{ flags map[string]models.Flag }

func (s fakeSrc) Get(_ context.Context, k string) (models.Flag, error) {
	f, ok := s.flags[k]
	if !ok {
		return models.Flag{}, store.ErrNotFound
	}
	return f, nil
}
func (s fakeSrc) List(context.Context) ([]models.Flag, error) {
	out := []models.Flag{}
	for _, f := range s.flags {
		out = append(out, f)
	}
	return out, nil
}

func newSrv() *Server {
	return New(fakeSrc{flags: map[string]models.Flag{
		"on":  {Key: "on", Enabled: true, Rollout: 100},
		"off": {Key: "off", Enabled: false, Rollout: 100},
	}}, nil, nil)
}

func TestEvaluate(t *testing.T) {
	s := newSrv()
	ctx := context.Background()

	if _, err := s.Evaluate(ctx, &flagcastv1.EvaluateRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Errorf("empty key: want InvalidArgument, got %v", err)
	}
	r, _ := s.Evaluate(ctx, &flagcastv1.EvaluateRequest{FlagKey: "missing", Context: &flagcastv1.Context{Key: "u"}})
	if r.GetValue() || r.GetReason() != "not_found" {
		t.Errorf("missing flag: got %+v", r)
	}
	r, _ = s.Evaluate(ctx, &flagcastv1.EvaluateRequest{FlagKey: "on", Context: &flagcastv1.Context{Key: "u"}})
	if !r.GetValue() || r.GetReason() != "enabled" {
		t.Errorf("on flag: got %+v", r)
	}
}

func TestEvaluateAll(t *testing.T) {
	r, err := newSrv().EvaluateAll(context.Background(), &flagcastv1.EvaluateAllRequest{Context: &flagcastv1.Context{Key: "u"}})
	if err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	if r.GetValues()["on"] != true || r.GetValues()["off"] != false {
		t.Errorf("values: %+v", r.GetValues())
	}
}

func TestWatchNilHub(t *testing.T) {
	// hub is nil → Watch unavailable.
	err := New(fakeSrc{}, nil, nil).Watch(&flagcastv1.WatchRequest{}, nil)
	if status.Code(err) != codes.Unavailable {
		t.Errorf("nil hub: want Unavailable, got %v", err)
	}
}
