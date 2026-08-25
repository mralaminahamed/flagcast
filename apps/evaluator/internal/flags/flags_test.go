package flags

import (
	"context"
	"testing"

	"github.com/mralaminahamed/flagcast/packages/shared/bus"
	"github.com/mralaminahamed/flagcast/packages/shared/models"
	"github.com/mralaminahamed/flagcast/packages/shared/store"
)

type fakeStore struct {
	flags map[string]models.Flag
	gets  int
}

func (s *fakeStore) Get(_ context.Context, k string) (models.Flag, error) {
	s.gets++
	f, ok := s.flags[k]
	if !ok {
		return models.Flag{}, store.ErrNotFound
	}
	return f, nil
}
func (s *fakeStore) List(_ context.Context) ([]models.Flag, error) {
	out := []models.Flag{}
	for _, f := range s.flags {
		out = append(out, f)
	}
	return out, nil
}

type fakeCache struct {
	m map[string]models.Flag
}

func newFakeCache() *fakeCache { return &fakeCache{m: map[string]models.Flag{}} }
func (c *fakeCache) Get(_ context.Context, k string) (models.Flag, bool, error) {
	f, ok := c.m[k]
	return f, ok, nil
}
func (c *fakeCache) GetAll(context.Context) ([]models.Flag, error) {
	out := []models.Flag{}
	for _, f := range c.m {
		out = append(out, f)
	}
	return out, nil
}
func (c *fakeCache) Set(_ context.Context, f models.Flag) error { c.m[f.Key] = f; return nil }
func (c *fakeCache) SetAll(_ context.Context, fs []models.Flag) error {
	c.m = map[string]models.Flag{}
	for _, f := range fs {
		c.m[f.Key] = f
	}
	return nil
}
func (c *fakeCache) Delete(_ context.Context, k string) error { delete(c.m, k); return nil }

func TestGetCacheThenStore(t *testing.T) {
	st := &fakeStore{flags: map[string]models.Flag{"a": {Key: "a", Enabled: true}}}
	c := newFakeCache()
	r := New(st, c)
	ctx := context.Background()

	// Miss → store → warms cache.
	if _, err := r.Get(ctx, "a"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if st.gets != 1 {
		t.Fatalf("expected 1 store get, got %d", st.gets)
	}
	// Hit → no further store read.
	if _, err := r.Get(ctx, "a"); err != nil || st.gets != 1 {
		t.Fatalf("expected cache hit (gets=1), got err=%v gets=%d", err, st.gets)
	}
	// Unknown propagates ErrNotFound.
	if _, err := r.Get(ctx, "nope"); err != store.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestOnChange(t *testing.T) {
	st := &fakeStore{flags: map[string]models.Flag{"a": {Key: "a", Enabled: true}}}
	c := newFakeCache()
	r := New(st, c)
	ctx := context.Background()

	// created/updated → reloaded into cache from the store.
	r.OnChange(ctx, bus.FlagChanged{Key: "a", Action: "updated"})
	if _, ok, _ := c.Get(ctx, "a"); !ok {
		t.Fatal("expected 'a' cached after update event")
	}
	// deleted → dropped from cache.
	r.OnChange(ctx, bus.FlagChanged{Key: "a", Action: "deleted"})
	if _, ok, _ := c.Get(ctx, "a"); ok {
		t.Fatal("expected 'a' removed after delete event")
	}
	// updated for a key the store no longer has → treated as delete.
	c.Set(ctx, models.Flag{Key: "ghost"})
	r.OnChange(ctx, bus.FlagChanged{Key: "ghost", Action: "updated"})
	if _, ok, _ := c.Get(ctx, "ghost"); ok {
		t.Fatal("expected 'ghost' dropped when missing from store")
	}
}

func TestNoCacheReadsStore(t *testing.T) {
	st := &fakeStore{flags: map[string]models.Flag{"a": {Key: "a"}}}
	r := New(st, nil) // no cache
	if _, err := r.Get(context.Background(), "a"); err != nil {
		t.Fatalf("get without cache: %v", err)
	}
}
