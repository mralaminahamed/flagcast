// Package flags is the evaluator's flag source: Mongo (source of truth) fronted
// by an optional Redis cache that is kept fresh from NATS flag-change events.
package flags

import (
	"context"
	"errors"

	"github.com/mralaminahamed/flagcast/packages/shared/bus"
	"github.com/mralaminahamed/flagcast/packages/shared/logger"
	"github.com/mralaminahamed/flagcast/packages/shared/metrics"
	"github.com/mralaminahamed/flagcast/packages/shared/models"
	"github.com/mralaminahamed/flagcast/packages/shared/store"
)

// Store is the flag source of truth (satisfied by *store.FlagStore).
type Store interface {
	Get(ctx context.Context, key string) (models.Flag, error)
	List(ctx context.Context) ([]models.Flag, error)
}

// Cache fronts the store (satisfied by *cache.FlagCache).
type Cache interface {
	Get(ctx context.Context, key string) (models.Flag, bool, error)
	GetAll(ctx context.Context) ([]models.Flag, error)
	Set(ctx context.Context, f models.Flag) error
	SetAll(ctx context.Context, flags []models.Flag) error
	Delete(ctx context.Context, key string) error
}

// Repo reads flags from the store, caching when a Cache is configured. Cache
// errors are logged and fall through to the store — the cache never breaks
// evaluation.
type Repo struct {
	store Store
	cache Cache // may be nil
}

func New(s Store, c Cache) *Repo { return &Repo{store: s, cache: c} }

// Warm loads every flag from Mongo into the cache at startup.
func (r *Repo) Warm(ctx context.Context) error {
	if r.cache == nil {
		return nil
	}
	flags, err := r.store.List(ctx)
	if err != nil {
		return err
	}
	return r.cache.SetAll(ctx, flags)
}

// Get returns a flag, preferring the cache. Propagates store.ErrNotFound.
func (r *Repo) Get(ctx context.Context, key string) (models.Flag, error) {
	if r.cache != nil {
		if f, ok, err := r.cache.Get(ctx, key); err != nil {
			logger.Log.Warn().Err(err).Msg("flag cache get; falling back to mongo")
		} else if ok {
			metrics.CacheOps.WithLabelValues("hit").Inc()
			return f, nil
		} else {
			metrics.CacheOps.WithLabelValues("miss").Inc()
		}
	}
	f, err := r.store.Get(ctx, key)
	if err != nil {
		return models.Flag{}, err
	}
	if r.cache != nil {
		if err := r.cache.Set(ctx, f); err != nil {
			logger.Log.Warn().Err(err).Msg("flag cache set")
		}
	}
	return f, nil
}

// List returns all flags, preferring the cache.
func (r *Repo) List(ctx context.Context) ([]models.Flag, error) {
	if r.cache != nil {
		if flags, err := r.cache.GetAll(ctx); err != nil {
			logger.Log.Warn().Err(err).Msg("flag cache getall; falling back to mongo")
		} else if len(flags) > 0 {
			return flags, nil
		}
	}
	return r.store.List(ctx)
}

// OnChange applies a flag-change event to the cache: reload from Mongo on
// create/update, drop on delete.
func (r *Repo) OnChange(ctx context.Context, evt bus.FlagChanged) {
	if r.cache == nil {
		return
	}
	if evt.Action == "deleted" {
		if err := r.cache.Delete(ctx, evt.Key); err != nil {
			logger.Log.Warn().Err(err).Str("flag", evt.Key).Msg("cache delete on change")
		}
		return
	}
	f, err := r.store.Get(ctx, evt.Key)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			_ = r.cache.Delete(ctx, evt.Key)
			return
		}
		logger.Log.Warn().Err(err).Str("flag", evt.Key).Msg("reload flag on change")
		return
	}
	if err := r.cache.Set(ctx, f); err != nil {
		logger.Log.Warn().Err(err).Str("flag", evt.Key).Msg("cache set on change")
	}
	logger.Log.Debug().Str("flag", evt.Key).Str("action", evt.Action).Msg("cache updated")
}
