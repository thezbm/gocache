package gocache

import (
	"github.com/thezbm/gocache/lru"
	"sync"
)

// cache is a thread-safe LRU cache.
type cache struct {
	mut      sync.Mutex
	lru      *lru.Cache
	capacity int64
}

// set stores a value in the cache with the given key.
// The LRU cache is lazy initialized.
func (c *cache) set(key string, value ByteView) {
	c.mut.Lock()
	defer c.mut.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.capacity, nil)
	}
	c.lru.Set(key, value)
}

// get retrieves a value from the cache by its key.
func (c *cache) get(key string) (ByteView, bool) {
	c.mut.Lock()
	defer c.mut.Unlock()
	if c.lru == nil {
		return ByteView{}, false
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return ByteView{}, false
}
