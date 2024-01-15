package SimpleCache

import (
	"SimpleCache/byteview"
	"SimpleCache/cache"
	"fmt"
	"log"
	"sync"
)

// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name      string
	getter    Getter
	mainCache cache.Cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// Get implements Getter interface function
// A callback function provided to the user, used to retrieve a value based on a key.
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache.Cache{CacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get value for a key from cache
func (g *Group) Get(key string) (byteview.ByteView, error) {
	if key == "" {
		return byteview.ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.Get(key); ok {
		log.Println("[SimpleCache] hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value byteview.ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (byteview.ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return byteview.ByteView{}, err

	}
	value := byteview.ByteView{B: byteview.CloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value byteview.ByteView) {
	g.mainCache.Add(key, value)
}
