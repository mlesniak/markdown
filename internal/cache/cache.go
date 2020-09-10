// A simple Cache abstraction. Is this actually feasible in Go or
// a we over-engineering a simple map structure?
package cache

// CacheEntry describes a Cache entry.
type Entry struct {
	Name string
	Data []byte
}

type Cache struct {
	cache map[string]Entry
}

func New() *Cache {
	ch := Cache{
		cache: make(map[string]Entry),
	}
	return &ch
}

func (c *Cache) Add(entry Entry) {
	c.cache[entry.Name] = entry
}

func (c *Cache) Get(name string) ([]byte, bool) {
	entry, ok := c.cache[name]
	if !ok {
		return nil, false
	}
	return entry.Data, true
}
