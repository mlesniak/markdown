// A simple Cache abstraction. Is this actually feasible in Go or
// a we over-engineering a simple map structure?
package cache

import "sync"

// CacheEntry describes a Cache entry.
type Entry struct {
	Name string
	Data []byte
}

type Cache struct {
	cache map[string]Entry
	lock  sync.Mutex
}

var once sync.Once
var singleton *Cache

func Get() *Cache {
	once.Do(func() {
		singleton = &Cache{
			cache: make(map[string]Entry),
		}
	})

	return singleton
}

func (c *Cache) AddEntry(entry Entry) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache[entry.Name] = entry
}

func (c *Cache) List() []string {
	keys := []string{}

	for k := range c.cache {
		keys = append(keys, k)
	}
	return keys
}

func (c *Cache) GetEntry(name string) ([]byte, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	entry, ok := c.cache[name]
	if !ok {
		return nil, false
	}
	return entry.Data, true
}
