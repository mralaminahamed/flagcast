// Package cache is the Redis-backed flag cache the evaluator reads from, kept in
// sync via NATS flag-change events. Mongo stays the source of truth.
package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mralaminahamed/flagcast/packages/shared/models"
	"github.com/redis/go-redis/v9"
)

const (
	flagsKey  = "flagcast:flags" // hash: field=flag key, value=flag JSON
	opTimeout = 3 * time.Second
)

type FlagCache struct {
	c *redis.Client
}

func NewFlagCache(ctx context.Context, url string) (*FlagCache, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	c := redis.NewClient(opt)
	if err := c.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &FlagCache{c: c}, nil
}

func (fc *FlagCache) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	return fc.c.Ping(ctx).Err()
}

func (fc *FlagCache) Close() error { return fc.c.Close() }

// Get returns a cached flag and whether it was present.
func (fc *FlagCache) Get(ctx context.Context, key string) (models.Flag, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	raw, err := fc.c.HGet(ctx, flagsKey, key).Result()
	if err == redis.Nil {
		return models.Flag{}, false, nil
	}
	if err != nil {
		return models.Flag{}, false, err
	}
	var f models.Flag
	if err := json.Unmarshal([]byte(raw), &f); err != nil {
		return models.Flag{}, false, nil // treat corrupt entry as a miss
	}
	return f, true, nil
}

// GetAll returns every cached flag.
func (fc *FlagCache) GetAll(ctx context.Context) ([]models.Flag, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	m, err := fc.c.HGetAll(ctx, flagsKey).Result()
	if err != nil {
		return nil, err
	}
	out := make([]models.Flag, 0, len(m))
	for _, v := range m {
		var f models.Flag
		if json.Unmarshal([]byte(v), &f) == nil {
			out = append(out, f)
		}
	}
	return out, nil
}

// Set writes/overwrites a flag in the cache.
func (fc *FlagCache) Set(ctx context.Context, f models.Flag) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	data, err := json.Marshal(f)
	if err != nil {
		return err
	}
	return fc.c.HSet(ctx, flagsKey, f.Key, data).Err()
}

// SetAll replaces the whole cache with the given flags (used to warm at startup).
func (fc *FlagCache) SetAll(ctx context.Context, flags []models.Flag) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	pipe := fc.c.TxPipeline()
	pipe.Del(ctx, flagsKey)
	for _, f := range flags {
		data, err := json.Marshal(f)
		if err != nil {
			return err
		}
		pipe.HSet(ctx, flagsKey, f.Key, data)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// Delete drops a flag from the cache.
func (fc *FlagCache) Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	return fc.c.HDel(ctx, flagsKey, key).Err()
}
