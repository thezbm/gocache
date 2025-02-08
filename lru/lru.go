package lru

import "container/list"

// An LRU cache.
type Cache struct {
	capacity int64                         // the maximum size of the cache; capacity <= 0 means no limit
	size     int64                         // the current size of the cache
	ll       *list.List                    // the underlying doubly linked list
	cache    map[string]*list.Element      // the key to element mapping
	onEvict  func(key string, value Value) // (optional) callback when an entry is evicted
}

// The element in the linked list. The KV pair of the cache.
type entry struct {
	key   string
	value Value
}

// A Value in the cache implements the Len method to return its size in bytes.
type Value interface {
	Len() int // the size in bytes
}

// The constructor of Cache.
func New(capacity int64, onEvict func(string, Value)) *Cache {
	return &Cache{
		capacity: capacity,
		ll:       list.New(),
		cache:    make(map[string]*list.Element),
		onEvict:  onEvict,
	}
}

// Get gets the value from the cache by key.
func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

// evict evicts LRU entry from the cache.
func (c *Cache) evict() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.size -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.onEvict != nil {
			c.onEvict(kv.key, kv.value)
		}
	}
}

// Set sets a value with a key in the cache.
func (c *Cache) Set(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.size += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.size += int64(len(key)) + int64(value.Len())
	}
	for c.capacity > 0 && c.size > c.capacity {
		c.evict()
	}
}

// Len returns the number of cache entries.
func (c *Cache) Len() int {
	return c.ll.Len()
}
