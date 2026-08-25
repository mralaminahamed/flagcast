// Package watch fans flag-change events out to connected Watch streams,
// re-evaluating each change for the individual watcher's context.
package watch

import (
	"context"
	"sync"

	"github.com/mralaminahamed/flagcast/packages/shared/bus"
	"github.com/mralaminahamed/flagcast/packages/shared/eval"
	flagcastv1 "github.com/mralaminahamed/flagcast/packages/shared/genproto/flagcast/v1"
	"github.com/mralaminahamed/flagcast/packages/shared/logger"
	"github.com/mralaminahamed/flagcast/packages/shared/models"
)

// FlagSource supplies the current flag for re-evaluation on change.
type FlagSource interface {
	Get(ctx context.Context, key string) (models.Flag, error)
}

type watcher struct {
	contextKey string
	ch         chan *flagcastv1.FlagChange
}

// Hub tracks active watchers and broadcasts per-context flag changes.
type Hub struct {
	src FlagSource
	mu  sync.Mutex
	m   map[int]*watcher
	seq int
}

// maxWatchers bounds concurrent Watch streams so an unauthenticated client can't
// exhaust memory/goroutines by opening unlimited streams.
const maxWatchers = 10000

func NewHub(src FlagSource) *Hub { return &Hub{src: src, m: map[int]*watcher{}} }

// Register adds a watcher for contextKey and returns its change channel plus an
// unregister func. ok is false when the watcher cap is reached. The channel is
// buffered and lossy: a slow watcher drops updates rather than blocking the
// broadcaster.
func (h *Hub) Register(contextKey string) (ch <-chan *flagcastv1.FlagChange, unregister func(), ok bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.m) >= maxWatchers {
		return nil, nil, false
	}
	id := h.seq
	h.seq++
	w := &watcher{contextKey: contextKey, ch: make(chan *flagcastv1.FlagChange, 16)}
	h.m[id] = w
	return w.ch, func() {
		h.mu.Lock()
		delete(h.m, id)
		h.mu.Unlock()
	}, true
}

// Broadcast re-evaluates the changed flag for every watcher and delivers a
// FlagChange. Call after the cache has been refreshed for the event.
func (h *Hub) Broadcast(ctx context.Context, evt bus.FlagChanged) {
	var f models.Flag
	deleted := evt.Action == "deleted"
	if !deleted {
		got, err := h.src.Get(ctx, evt.Key)
		if err != nil {
			// If it's gone, treat the change as a deletion for watchers.
			deleted = true
		} else {
			f = got
		}
	}

	h.mu.Lock()
	watchers := make([]*watcher, 0, len(h.m))
	for _, w := range h.m {
		watchers = append(watchers, w)
	}
	h.mu.Unlock()

	for _, w := range watchers {
		fc := &flagcastv1.FlagChange{FlagKey: evt.Key, Deleted: deleted}
		if !deleted {
			v, reason := eval.Evaluate(f, w.contextKey)
			fc.Value, fc.Reason = v, reason
		}
		select {
		case w.ch <- fc:
		default:
			logger.Log.Warn().Str("flag", evt.Key).Msg("watch: dropping update to slow client")
		}
	}
}
