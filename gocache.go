package gocache

import (
	"fmt"
	"log"
	"sync"
)

// A Getter loads data in bytes with a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// A Group is a cache namespace and associated data loaded spread over one or more nodes.
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup creates a new instance of Group.
// A 0 capacity means no limit of the cache size.
func NewGroup(name string, capacity int64, getter Getter) *Group {
	if getter == nil {
		panic("getter is nil")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{capacity: capacity},
	}
	groups[name] = g
	return g
}

// GetGroup returns the Group instance by name.
// If the group does not exist, it returns nil.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get gets the value for the given key from the cache.
// If the key does not exist, it loads the value.
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Printf("[gocache] hit with key=%s", key)
		return v, nil
	}
	return g.load(key)
}

// Load loads the value either from its peers or from the local node by calling the getter.
func (g *Group) load(key string) (ByteView, error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			value, err := g.getFromPeer(peer, key)
			if err == nil {
				return value, nil
			}
			log.Println("[gocache] failed to get from peer", err)
		}
	}
	return g.getLocally(key)
}

// getLocally loads the value using the getter and stores it in the cache.
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{bytes: copyBytes(bytes)}
	g.populateCache(key, value)
	log.Printf("[gocache] load with key=%s", key)
	return value, nil
}

// populateCache stores the value in the cache of the group.
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.set(key, value)
}

// getFromPeer retrieves the value from the peer.
func (g *Group) getFromPeer(peer Peer, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{bytes: bytes}, nil
}

// RegisterPeers registers a PeerPicker for choosing remote peers.
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("[gocache] RegisterPeers called more than once")
	}
	g.peers = peers
}
