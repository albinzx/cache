package memory

import (
	"context"
	"time"

	"github.com/albinzx/cache"
	mem "github.com/patrickmn/go-cache"
)

// Cacher is cache implementation using memory
type Cacher struct {
	cache *mem.Cache
	ttl   time.Duration
}

// defaults sets default cacher option
func defaults(cacher *Cacher) {
	if cacher.cache == nil {
		if cacher.ttl > time.Second {
			cacher.cache = mem.New(cacher.ttl, 10*time.Minute)
		} else {
			cacher.cache = mem.New(mem.NoExpiration, 10*time.Minute)
		}
	}
}

// Option provides cacher options
type Option func(*Cacher)

// New returns new memory cacher
func New(options ...Option) *Cacher {
	mcache := &Cacher{}

	for _, option := range options {
		option(mcache)
	}

	defaults(mcache)

	return mcache
}

func (c *Cacher) Set(ctx context.Context, key string, value any, setOptions ...cache.SetOption) error {
	setConfig := &cache.SetConfiguration{
		TTL: c.ttl}
	for _, option := range setOptions {
		option(setConfig)
	}

	c.cache.Set(key, value, setConfig.TTL)

	return nil
}

func (c *Cacher) Get(ctx context.Context, key string) (any, error) {
	if value, ok := c.cache.Get(key); ok {
		return value, nil
	}

	return nil, nil
}

func (c *Cacher) Delete(ctx context.Context, key string) error {
	c.cache.Delete((key))

	return nil
}

func (c *Cacher) Load(ctx context.Context, data map[string]any) error {
	for key, val := range data {
		c.cache.Set(key, val, c.ttl)
	}

	return nil
}

func (c *Cacher) Close() error {
	c.cache.Flush()
	return nil
}

// WithTTL returns option to set global TTL
func WithTTL(ttl time.Duration) Option {
	return func(cache *Cacher) {
		cache.ttl = ttl
	}
}
