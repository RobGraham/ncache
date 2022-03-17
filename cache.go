package ncache

import (
	"fmt"
	"sync"
	"time"
)

type entry struct {
	value      interface{}
	expiration int64
}

type Cache struct {
	mu sync.Mutex
	// Cache store
	entries *sync.Map
	// Callback when cache is evicted/deleted
	onEvict func(k string, v interface{})
	// Evict interval timer
	evict time.Duration

	// TODO: optional delivery of metrics
	// hits         uint64
	// misses       uint64
	// totalEntries uint64
}

type Config struct {
	// Interval that the cache should conduct eviction
	// of stale cache entries. If not passed, or set to
	// 0, there is no eviction policy
	Evict time.Duration
	// Optional callback to be executed when cache entries
	// are evicted or deleted
	OnEvict func(key, value interface{})
}

// New returns an instance of `Cache` and any configuration errors
func New(config *Config) (*Cache, error) {
	if config == nil {
		return nil, fmt.Errorf("must pass a configuration")
	}

	cache := &Cache{
		entries: &sync.Map{},
		evict:   config.Evict,
	}

	cache.onEvict = func(k string, v interface{}) {
		if config.OnEvict != nil {
			config.OnEvict(k, v)
		}
	}

	go cache.evictor()

	return cache, nil
}

// Add an entry to the cache only if it doesn't already exist for the given
// key otherwise returns an error. If the duration is 0 the item never expires.
func (c *Cache) Add(k string, v interface{}, ttl time.Duration) error {
	if _, found := c.Get(k); found {
		return fmt.Errorf("entry %s already exists", k)
	}

	c.Set(k, v, ttl)

	return nil
}

// Add an entry to the cache, replacing any existing item.
// If the duration is 0 the item never expires.
func (c *Cache) Set(k string, v interface{}, ttl time.Duration) {
	var e int64

	if ttl > 0 {
		e = time.Now().Add(ttl).UnixNano()
	}

	c.entries.Store(k, &entry{
		value:      v,
		expiration: e,
	})
}

// Get an entry from the cache. Returns the entry or nil, and a bool
// indicating whether the key was found.
func (c *Cache) Get(k string) (interface{}, bool) {
	v, found := c.entries.Load(k)
	if !found {
		return nil, false
	}

	entry := v.(*entry)

	if entry.expiration > 0 {
		if time.Now().UnixNano() > entry.expiration {
			return nil, false
		}
	}

	return entry.value, true
}

// Deletes the entry from the cache if it exists.
func (c *Cache) Delete(k string) {
	if e, found := c.entries.LoadAndDelete(k); found {
		c.onEvict(k, e.(*entry).value)
	}
}

// Flush removes all cache entries. Use with caution.
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = &sync.Map{}
}

// Creates up an interval if `evict` is defined in
// which it loops over all cache entries and checks
// their TTL against the current time and removes
// any stale entries.
func (c *Cache) evictor() {
	if c.evict > 0 {
		ticker := time.NewTicker(c.evict)

		for range ticker.C {
			c.entries.Range(func(k, v interface{}) bool {
				if v.(*entry).expiration < time.Now().UnixNano() {
					c.Delete(k.(string))
				}
				return true
			})
		}
	}
}
