package i18nmod

import (
	"sync"

	"gopkg.in/fatih/set.v0"
)

type Cache struct {
	Data    map[string]map[string]*Result
	Allowed set.Interface
	*sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{Data: make(map[string]map[string]*Result), Allowed: set.New(set.ThreadSafe)}
}

func (c *Cache) Allow(key string) *Cache {
	c.Allowed.Add(key)
	return c
}

func (c *Cache) AddIfAllowed(t *T, lang string, r *Result) *Cache {
	if c.Allowed.Has(t.Key) {
		defer c.Unlock()
		c.Lock()
		if _, ok := c.Data[lang]; !ok {
			c.Data[lang] = make(map[string]*Result)
		}
		c.Data[lang][t.Key.Key] = r
	}
	return c
}
