package cache

import (
	"SimpleCache/cache/singleflight"
	"SimpleCache/common/byteview"
	pb "SimpleCache/common/pb"
	"SimpleCache/common/peer"
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
	mainCache Cache
	peers     peer.PeerPicker
	// use singleflight.Group to make sure that
	// each key is only fetched once
	loader *singleflight.Group
}

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers peer.PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
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
		mainCache: Cache{CacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
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
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[SimpleCache] Failed to get from peer", err)
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(byteview.ByteView), nil
	}
	return
}

func (g *Group) getFromPeer(peer peer.PeerGetter, key string) (byteview.ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return byteview.ByteView{}, err
	}
	return byteview.ByteView{B: res.Value}, nil

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
