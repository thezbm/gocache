package singleflight

import "sync"

// A call is function call that is either in-flight or completed.
type call struct {
	wg  sync.WaitGroup // syncs multiple calls
	val any            // the returned value
	err error
}

// A Group is a collection of calls with distinct keys.
type Group struct {
	mu sync.Mutex
	m  map[string]*call // maps keys to calls
}

// Do wraps a function call to ensure that only one call is made at a time for a given key.
func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	// If the key is being tracked, use the existing call directly.
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	// Otherwise, create a new call and track the key.
	c := new(call)
	g.m[key] = c
	// The WaitGroup Add should be in the critical section.
	// Executing concurrently with WaitGroup Wait causes data race.
	c.wg.Add(1)
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key) // untrack the key
	g.mu.Unlock()

	return c.val, c.err
}
