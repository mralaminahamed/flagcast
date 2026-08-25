package watch

import (
	"context"
	"testing"

	"github.com/mralaminahamed/flagcast/packages/shared/bus"
	"github.com/mralaminahamed/flagcast/packages/shared/models"
)

type fakeSource struct{ f models.Flag }

func (s fakeSource) Get(context.Context, string) (models.Flag, error) { return s.f, nil }

func TestRegisterCap(t *testing.T) {
	h := NewHub(fakeSource{})
	unregisters := make([]func(), 0, maxWatchers)
	for i := 0; i < maxWatchers; i++ {
		_, un, ok := h.Register("u")
		if !ok {
			t.Fatalf("register %d unexpectedly refused", i)
		}
		unregisters = append(unregisters, un)
	}
	if _, _, ok := h.Register("u"); ok {
		t.Fatal("register past cap should be refused")
	}
	// Freeing a slot lets a new watcher in.
	unregisters[0]()
	if _, _, ok := h.Register("u"); !ok {
		t.Fatal("register after unregister should succeed")
	}
}

func TestBroadcastReEvaluatesPerWatcher(t *testing.T) {
	h := NewHub(fakeSource{f: models.Flag{Key: "k", Enabled: true, Rollout: 100}})
	ch, _, ok := h.Register("u1")
	if !ok {
		t.Fatal("register failed")
	}
	h.Broadcast(context.Background(), bus.FlagChanged{Key: "k", Action: "updated"})
	select {
	case fc := <-ch:
		if fc.GetFlagKey() != "k" || !fc.GetValue() {
			t.Fatalf("got %+v, want k=true", fc)
		}
	default:
		t.Fatal("no FlagChange delivered")
	}
}

func TestBroadcastDropsInsteadOfBlocking(t *testing.T) {
	h := NewHub(fakeSource{f: models.Flag{Key: "k", Enabled: true, Rollout: 100}})
	_, _, ok := h.Register("u1")
	if !ok {
		t.Fatal("register failed")
	}
	// Never drain the channel; more broadcasts than the 16 buffer must not block.
	for i := 0; i < 100; i++ {
		h.Broadcast(context.Background(), bus.FlagChanged{Key: "k", Action: "updated"})
	}
}
