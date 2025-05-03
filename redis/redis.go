package redis

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/albinzx/cache"
	"github.com/albinzx/cache/internal"
	"github.com/albinzx/marshal"
	goredis "github.com/redis/go-redis/v9"
)

// Cacher is cache implementation with redis
type Cacher struct {
	client      goredis.UniversalClient
	ttl         time.Duration
	prefix      internal.KeyPrefix
	marshaller  marshal.Marshaller
	closeClient bool
}

// defaults sets default redis cacher option
func defaults(cache *Cacher) {
	if cache.client == nil {
		cache.client = goredis.NewClient(&goredis.Options{})
	}

	if cache.prefix == nil {
		cache.prefix = &internal.NoPrefix{}
	}
}

// Option provides redis cacher options
type Option func(*Cacher)

// New returns new redis cacher
func New(options ...Option) *Cacher {
	rcache := &Cacher{closeClient: true}

	for _, option := range options {
		option(rcache)
	}

	defaults(rcache)

	return rcache
}

func (c *Cacher) Set(ctx context.Context, key string, value any, setOptions ...cache.SetOption) error {
	setConfig := &cache.SetConfiguration{
		TTL: c.ttl}
	for _, option := range setOptions {
		option(setConfig)
	}

	if c.marshaller != nil {
		// if marshaller is set, marshal value
		// before storing to redis
		marshalled, err := c.marshaller.Marshal(value)
		if err != nil {
			return err
		}
		value = marshalled
	}

	return c.client.Set(ctx, c.prefix.Prefix(key), value, setConfig.TTL).Err()
}

func (c *Cacher) Get(ctx context.Context, key string) (any, error) {
	value := c.client.Get(ctx, c.prefix.Prefix(key))

	if errors.Is(value.Err(), goredis.Nil) {
		return nil, nil
	}

	if c.marshaller != nil {
		// if marshaller is set, unmarshal value
		marshalled, err := value.Bytes()
		if err != nil {
			return nil, err
		}
		unmarshalled, err := c.marshaller.Unmarshal(marshalled)
		if err != nil {
			return nil, err
		}
		return unmarshalled, nil
	}

	// if marshaller is not set, return value as string
	return value.Val(), nil
}

func (c *Cacher) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.prefix.Prefix(key)).Err()
}

func (c *Cacher) Load(ctx context.Context, data map[string]any) error {

	if c.marshaller != nil {
		// if marshaller is set, marshal all values
		// before storing to redis
		bytesMap := make(map[string][]byte, len(data))

		for key, val := range data {
			marshalled, err := c.marshaller.Marshal(val)
			if err == nil {
				bytesMap[key] = marshalled
			}
		}

		_, err := c.client.Pipelined(ctx, func(pipe goredis.Pipeliner) error {

			for key, val := range bytesMap {
				pipe.Set(ctx, c.prefix.Prefix(key), val, c.ttl)
			}

			return nil
		})

		return err
	}

	// if marshaller is not set, store values as is to redis
	_, err := c.client.Pipelined(ctx, func(pipe goredis.Pipeliner) error {

		for key, val := range data {
			pipe.Set(ctx, c.prefix.Prefix(key), val, c.ttl)
		}

		return nil
	})

	return err
}

func (c *Cacher) Close() error {
	if c.closeClient {
		return c.client.Close()
	}

	return nil
}

// WithRedisClient returns option with redis client
func WithRedisClient(client goredis.UniversalClient) Option {
	return func(cache *Cacher) {
		cache.client = client
	}
}

// WithSharedRedisClient returns option with shared redis client
// if closeClient is true, redis client will be closed when this cacher is closed
// among cachers that use same shared redis client, makes sure only one cacher is set closeClient to true
func WithSharedRedisClient(client goredis.UniversalClient, closeClient bool) Option {
	return func(cache *Cacher) {
		cache.client = client
		cache.closeClient = closeClient
	}
}

// WithTTL returns option to set global TTL
func WithTTL(ttl time.Duration) Option {
	return func(cache *Cacher) {
		cache.ttl = ttl
	}
}

// WithName returns option to add name as prefix to key
// if name is empty, no prefix will be added
func WithName(name string) Option {
	return func(cache *Cacher) {
		var keyPrefix internal.KeyPrefix
		if len(name) == 0 {
			keyPrefix = &internal.NoPrefix{}
		} else {
			// prefix is always in lower case
			keyPrefix = &internal.WithPrefix{Name: strings.ToLower(name)}
		}

		cache.prefix = keyPrefix
	}
}

// WithMarshaller returns option to set marshaller
func WithMarshaller(marshaller marshal.Marshaller) Option {
	return func(cache *Cacher) {
		cache.marshaller = marshaller
	}
}
