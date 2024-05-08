package cache

import (
	"fmt"
	"sync"
)

type ErrKeyNotFound struct {
	key string
	error
}

func (e ErrKeyNotFound) Error() string {
	return fmt.Sprintf("key %s not found in cache", e.key)
}

type ErrKeyExpired struct {
	key string
	error
}

func (e ErrKeyExpired) Error() string {
	return fmt.Sprintf("key %s has expired", e.key)
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]*Item
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]*Item),
	}
}

func (c *MemoryCache) Get(key KeyGetter) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keyGetter := key
	item, ok := c.items[keyGetter.GetKey()]
	if !ok {
		return nil, ErrKeyNotFound{key: keyGetter.GetKey()}
	}
	if !item.Expired() {
		return item.Value(), nil
	}
	delete(c.items, keyGetter.GetKey())
	return nil, ErrKeyExpired{key: keyGetter.GetKey()}
}

func (c *MemoryCache) Set(key KeyGetter, value any, options ...ItemOption) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := NewItem(value)

	for _, option := range options {
		err := option(item)
		if err != nil {
			return err
		}
	}

	keyGetter := key
	c.items[keyGetter.GetKey()] = item
	return nil
}

func (c *MemoryCache) Delete(key KeyGetter) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key.GetKey())
	return nil
}

func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*Item)
	return nil
}
