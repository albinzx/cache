package cache

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	// ErrCacherNil is returned when cacher is nil
	ErrCacherNil = errors.New("cacher is nil")
)

// Cache defines cache operation
type Cache interface {
	// Set stores or replaces key-value to cache
	Set(context.Context, string, any, ...SetOption) error
	// Get retrieves value from cache
	Get(context.Context, string) (any, error)
	// Delete deletes value from cache
	Delete(context.Context, string) error
}

// SetConfiguration holds configuration for set operation
type SetConfiguration struct {
	TTL time.Duration
}

// SetOption provides options for set operation
type SetOption func(setConfig *SetConfiguration)

// WithTTL sets cache time to live, this option override global cache TTL
func WithTTL(ttl time.Duration) SetOption {
	return func(setConfig *SetConfiguration) {
		setConfig.TTL = ttl
	}
}

// Cacher defines operation for cache implementation
type Cacher interface {
	io.Closer
	// Set sets key-value to cache
	Set(context.Context, string, any, ...SetOption) error
	// Get gets value from cache
	Get(context.Context, string) (any, error)
	// Delete deletes value from cache
	Delete(context.Context, string) error
	// Load loads multiple key-values into cache
	Load(context.Context, map[string]any) error
}

// Persister defines operation for persistence storage
// Persister is used to persist cache data to storage
// and load it back to cache when needed
type Persister interface {
	io.Closer
	// Save stores key value to persistence storage
	Save(ctx context.Context, key string, value any) error
	// SelectOne retrieves value by key from persistence storage
	SelectOne(ctx context.Context, key string) (any, error)
	// SelectAll retrieves all key-values from persistence storage
	SelectAll(ctx context.Context) (map[string]any, error)
	// Delete deletes value by key from persistence storage
	Delete(ctx context.Context, key string) error
}

// PatternedCache defines caching pattern
type PatternedCache struct {
	cacher    Cacher
	persister Persister
	pattern   Pattern
}

// New creates a new cache with the given cacher and persister
func New(cacher Cacher, persister Persister, options ...Option) (*PatternedCache, error) {
	if cacher == nil {
		return nil, ErrCacherNil
	}

	cache := &PatternedCache{
		cacher:    cacher,
		persister: persister,
	}

	for _, option := range options {
		option(cache)
	}
	defaults(cache)

	return cache, nil
}

// Option provides cache options
type Option func(c *PatternedCache)

// defaults sets default cache option
func defaults(c *PatternedCache) {
	if c.pattern == nil {
		c.pattern = &CacheAside{}
	}
}

// WithPattern returns option to set cache pattern
func WithPattern(pattern Pattern) Option {
	return func(c *PatternedCache) {
		c.pattern = pattern
	}
}

// Set sets key-value to cache
func (c *PatternedCache) Set(ctx context.Context, key string, value any, options ...SetOption) error {
	return c.pattern.Set(ctx, key, value, c.cacher, c.persister, options...)
}

// Get retrieves value from cache
func (c *PatternedCache) Get(ctx context.Context, key string) (any, error) {
	return c.pattern.Get(ctx, key, c.cacher, c.persister)
}

// Delete deletes value from cache
func (c *PatternedCache) Delete(ctx context.Context, key string) error {
	return c.pattern.Delete(ctx, key, c.cacher, c.persister)
}
