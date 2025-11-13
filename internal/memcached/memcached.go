package memcached

import (
	"context"
	"time"

	"github.com/Caritas-Team/reviewer/internal/config"
	"github.com/bradfitz/gomemcache/memcache"
)

var ErrCacheMiss = memcache.ErrCacheMiss

type Cache struct {
	client *memcache.Client
	ttl    time.Duration
	prefix string
	enable bool
}

type CacheInterface interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Close() error
}

func NewCache(ctx context.Context, cfg config.Config) (*Cache, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !cfg.Memcached.Enable {
		return &Cache{enable: false}, nil
	} else {
		client := memcache.New(cfg.Memcached.Servers...)

		if err := client.Ping(); err != nil {
			return nil, err
		}

		return &Cache{
			client: client,
			ttl:    time.Duration(cfg.Memcached.DefaultTTL) * time.Second,
			prefix: cfg.Memcached.KeyPrefix,
			enable: true,
		}, nil
	}
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !c.enable || c.client == nil {
		return nil, memcache.ErrCacheMiss
	}
	prefix := c.prefix + ":" + key
	item, err := c.client.Get(prefix)
	if err != nil {
		return nil, err
	}
	return item.Value, nil
}

func (c *Cache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !c.enable || c.client == nil {
		return memcache.ErrCacheMiss
	}
	prefix := c.prefix + ":" + key
	err := c.client.Set(&memcache.Item{
		Key:        prefix,
		Value:      value,
		Expiration: int32(ttl.Seconds()),
	})
	return err
}

func (c *Cache) Close() error {
	return c.client.Close()
}
