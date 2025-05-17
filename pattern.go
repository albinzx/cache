package cache

import (
	"context"
	"log"
)

type Pattern interface {
	Set(context.Context, string, any, Cacher, Persister, ...SetOption) error
	Get(context.Context, string, Cacher, Persister) (any, error)
	Delete(context.Context, string, Cacher, Persister) error
}

// CacheAside is a cache pattern that works along side with persistence storage
// operation on persistence storage is handled by the caller
// if you want automatic read/write on the cache and persistence storage, use other patterns
type CacheAside struct {
}

// Set stores key-value to cache
func (r *CacheAside) Set(ctx context.Context, key string, value any, c Cacher, _ Persister, options ...SetOption) error {
	if err := c.Set(ctx, key, value, options...); err != nil {
		return err
	}

	return nil
}

// Get retrieves value from cache
func (r *CacheAside) Get(ctx context.Context, key string, c Cacher, _ Persister) (any, error) {
	return c.Get(ctx, key)
}

// Delete deletes value from cache
func (r *CacheAside) Delete(ctx context.Context, key string, c Cacher, _ Persister) error {
	if err := c.Delete(ctx, key); err != nil {
		return err
	}

	return nil
}

// ReadThrough is a cache pattern that reads from cache first
// and if not found, reads from persistence storage
// and stores the value to cache
type ReadThrough struct {
}

// Set stores key-value to cache
func (r *ReadThrough) Set(ctx context.Context, key string, value any, c Cacher, p Persister, options ...SetOption) error {
	if err := c.Set(ctx, key, value, options...); err != nil {
		return err
	}

	return nil
}

// Get retrieves value from cache
// if not found, retrieves value from persistence storage
// and stores the value to cache
// if value is nil, it means the key is not found in both cache and persistence storage
func (r *ReadThrough) Get(ctx context.Context, key string, c Cacher, p Persister) (any, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		log.Printf("failed to get value to cache: %v", err)
	}

	if value == nil && p != nil {
		value, err = p.SelectOne(ctx, key)
		if err != nil {
			return nil, err
		}

		if value != nil {
			if err := c.Set(ctx, key, value); err != nil {
				log.Printf("failed to set value to cache: %v", err)
			}
		}
	}

	return value, nil
}

// Delete deletes value from cache
func (r *ReadThrough) Delete(ctx context.Context, key string, c Cacher, p Persister) error {
	if err := c.Delete(ctx, key); err != nil {
		return err
	}

	return nil
}

// WriteThrough is a cache pattern that writes to cache first and then writes to persistence storage
// if persistence storage is not available, it will only write to cache
type WriteThrough struct {
}

// Set stores key-value to cache and persistence storage
func (w *WriteThrough) Set(ctx context.Context, key string, value any, c Cacher, p Persister, options ...SetOption) error {
	if err := c.Set(ctx, key, value, options...); err != nil {
		return err
	}

	if p != nil {
		if err := p.Save(ctx, key, value); err != nil {
			log.Printf("failed to save value to persistence storage: %v", err)

			if derr := c.Delete(ctx, key); derr != nil {
				log.Printf("failed to delete value from cache: %v", derr)
			}

			return err
		}
	}

	return nil
}

// Get retrieves value from cache
// if not found, retrieves value from persistence storage
// and stores the value to cache
func (w *WriteThrough) Get(ctx context.Context, key string, c Cacher, p Persister) (any, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		log.Printf("failed to get value to cache: %v", err)
	}

	if value == nil && p != nil {
		value, err = p.SelectOne(ctx, key)
		if err != nil {
			return nil, err
		}

		if value != nil {
			if err := c.Set(ctx, key, value); err != nil {
				log.Printf("failed to set value to cache: %v", err)
			}
		}
	}

	return value, nil
}

// Delete deletes value from cache and persistence storage
func (w *WriteThrough) Delete(ctx context.Context, key string, c Cacher, p Persister) error {
	if err := c.Delete(ctx, key); err != nil {
		return err
	}

	if p != nil {
		if err := p.Delete(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

// WriteBehind is a cache pattern that writes to cache first
// and then writes to persistence storage asynchronously
type WriteBehind struct {
}

// Set stores key-value to cache and asynchronously to persistence storage
func (w *WriteBehind) Set(ctx context.Context, key string, value any, c Cacher, p Persister, options ...SetOption) error {
	if err := c.Set(ctx, key, value, options...); err != nil {
		return err
	}

	if p != nil {
		go func() {
			if err := p.Save(ctx, key, value); err != nil {
				log.Printf("failed to save value to persistence storage: %v", err)

				if derr := c.Delete(ctx, key); derr != nil {
					log.Printf("failed to delete value from cache: %v", derr)
				}
			}
		}()
	}

	return nil
}

// Get retrieves value from cache
// if not found, retrieves value from persistence storage
// and stores the value to cache
func (w *WriteBehind) Get(ctx context.Context, key string, c Cacher, p Persister) (any, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		log.Printf("failed to get value to cache: %v", err)
	}

	if value == nil && p != nil {
		value, err = p.SelectOne(ctx, key)
		if err != nil {
			return nil, err
		}

		if value != nil {
			if err := c.Set(ctx, key, value); err != nil {
				log.Printf("failed to set value to cache: %v", err)
			}
		}
	}

	return value, nil
}

// Delete deletes value from cache and asynchronously from persistence storage
func (w *WriteBehind) Delete(ctx context.Context, key string, c Cacher, p Persister) error {
	if err := c.Delete(ctx, key); err != nil {
		return err
	}

	if p != nil {
		go func() {
			if err := p.Delete(ctx, key); err != nil {
				log.Printf("failed to delete value from persistence storage: %v", err)
			}
		}()
	}

	return nil
}

// WriteAround is a cache pattern that writes to persistence storage but not to cache
// write to cache is done with lazy loading on read
type WriteAround struct {
}

// Set stores key-value to persistence storage
func (w *WriteAround) Set(ctx context.Context, key string, value any, _ Cacher, p Persister, _ ...SetOption) error {
	if p != nil {
		if err := p.Save(ctx, key, value); err != nil {
			log.Printf("failed to save value to persistence storage: %v", err)

			return err
		}
	}

	return nil
}

// Get retrieves value from cache
// if not found, retrieves value from persistence storage
// and stores the value to cache
func (w *WriteAround) Get(ctx context.Context, key string, c Cacher, p Persister) (any, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		log.Printf("failed to get value to cache: %v", err)
	}

	if value == nil && p != nil {
		value, err = p.SelectOne(ctx, key)
		if err != nil {
			return nil, err
		}

		if value != nil {
			if err := c.Set(ctx, key, value); err != nil {
				log.Printf("failed to set value to cache: %v", err)
			}
		}
	}

	return value, nil
}

// Delete deletes value from persistence storage and cache
func (w *WriteAround) Delete(ctx context.Context, key string, c Cacher, p Persister) error {
	if p != nil {
		if err := p.Delete(ctx, key); err != nil {
			log.Printf("failed to delete value from persistence storage: %v", err)

			return err
		}
	}

	if err := c.Delete(ctx, key); err != nil {
		return err
	}

	return nil
}
