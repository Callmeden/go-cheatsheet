// Package cache implements a simple in-memory cache
// with Set, Get, and Delete methods.
package cache

type Cache struct {
	data map[string]any
}

func New() *Cache {
	return &Cache{data: make(map[string]any)}
}

func (c *Cache) Set(key string, value any) {
	c.data[key] = value
}

func (c *Cache) Get(key string) (any, bool) {
	data, exists := c.data[key]
	return data, exists
}

func (c *Cache) Delete(key string) {
	delete(c.data, key)
}
