package yokaman

import (
	"errors"
	"sync"
)

type Cache struct {
	cache map[string]uint8
	mu    sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		cache: make(map[string]uint8),
	}
}

func (c *Cache) Set(key string, value uint8) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = value
}

func (c *Cache) Get(key string) (uint8, error) {
	/*c.mu.RLock()
	defer c.mu.RUnlock()*/

	value, _ := c.cache[key]
	if _, ok := c.cache[key]; ok {
		return value, nil
	} else {
		//fmt.Println("key 'a' does not exist")
		return 0xff, errors.New("element not exist")
	}

}
