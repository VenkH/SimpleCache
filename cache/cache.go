package cache

import (
	"SimpleCache/cache/lru"
	"SimpleCache/common/byteview"
	"sync"
)

// Cache Internal-facing Cache
type Cache struct {
	Mu         sync.Mutex
	Lru        *lru.LruCache
	CacheBytes int64
}

func (c *Cache) Add(key string, value byteview.ByteView) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if c.Lru == nil {
		c.Lru = lru.New(c.CacheBytes, nil)
	}
	c.Lru.Add(key, value)
}

func (c *Cache) Get(key string) (value byteview.ByteView, ok bool) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if c.Lru == nil {
		return
	}

	if v, ok := c.Lru.Get(key); ok {
		return v.(byteview.ByteView), ok
	}

	return
}
