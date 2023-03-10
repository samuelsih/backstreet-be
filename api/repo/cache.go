package repo

import (
	"backstreetlinkv2/api/helper"
	"context"
	"errors"
	"github.com/allegro/bigcache/v3"
	"github.com/rs/zerolog/log"
	"time"
)

var (
	CacheNotFound = errors.New("not found")
)

type Cache struct {
	provider *bigcache.BigCache
}

func NewCache(ctx context.Context, duration time.Duration) (*Cache, error) {
	cache, err := bigcache.New(ctx, bigcache.Config{
		Shards:             1024,
		LifeWindow:         duration,
		CleanWindow:        1 * time.Minute,
		MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntrySize:       500,
		StatsEnabled:       false,
		Verbose:            true,
		HardMaxCacheSize:   30,
		OnRemoveWithReason: func(key string, entry []byte, reason bigcache.RemoveReason) {
			switch reason {
			case bigcache.Expired:
				log.Info().Msgf("CACHE: %s has been removed because [expired]", key)
			case bigcache.NoSpace:
				log.Info().Msgf("CACHE: %s has been removed because [no space]", key)
			case bigcache.Deleted:
				log.Info().Msgf("CACHE: %s has been removed because [deleted]", key)
			}
		},
	})
	if err != nil {
		return nil, err
	}

	return &Cache{
		provider: cache,
	}, nil
}

func (c *Cache) Get(key string) ([]byte, error) {
	const op = helper.Op("Cache.Set")

	value, err := c.provider.Get(key)
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			return nil, CacheNotFound
		}

		return nil, helper.E(helper.Op("Cache.Set"), helper.KindUnexpected, err, "can't get the value")
	}

	return value, nil
}

func (c *Cache) Set(key string, value []byte) error {
	if err := c.provider.Set(key, value); err != nil {
		return helper.E(helper.Op("Cache.Set"), helper.GetKind(err), err, "can't store in server")
	}

	return nil
}
