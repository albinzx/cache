package cache

import "context"

type Pattern interface {
	Set(context.Context, string, any, Cacher, Persister) error
	Get(context.Context, string, Cacher, Persister) (any, error)
	Delete(context.Context, string, Cacher, Persister) error
}

// ReadAside is a cache pattern that reads from cache first
// and if not found, reads from persistence storage
// and stores the value to cache
type ReadAside struct {
}

// Set stores key-value to cache
func (r *ReadAside) Set(ctx context.Context, key string, value any, c Cacher, p Persister) error {
	if err := c.Set(ctx, key, value); err != nil {
		return err
	}

	return nil
}

// Get retrieves value from cache
// if not found, retrieves value from persistence storage
// and stores the value to cache
// if value is nil, it means the key is not found in both cache and persistence storage
func (r *ReadAside) Get(ctx context.Context, key string, c Cacher, p Persister) (any, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if value == nil {
		value, err = p.SelectOne(ctx, key)
		if err != nil {
			return nil, err
		}

		if value != nil {
			if err := c.Set(ctx, key, value); err != nil {
				return value, err
			}
		}
	}

	return value, nil
}

// Delete deletes value from cache
func (r *ReadAside) Delete(ctx context.Context, key string, c Cacher, p Persister) error {
	if err := c.Delete(ctx, key); err != nil {
		return err
	}

	return nil
}
