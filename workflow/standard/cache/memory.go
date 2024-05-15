// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package cache

import (
	"fmt"
	"sync"
)

// ErrKeyNotFound represents an error indicating that a key was not found in the cache.
type ErrKeyNotFound struct {
	key string
	error
}

// Error returns the error message for ErrKeyNotFound.
func (e ErrKeyNotFound) Error() string {
	return fmt.Sprintf("key %s not found in cache", e.key)
}

// ErrKeyExpired represents an error indicating that a key has expired in the cache.
type ErrKeyExpired struct {
	key string
	error
}

// Error returns the error message for ErrKeyExpired.
func (e ErrKeyExpired) Error() string {
	return fmt.Sprintf("key %s has expired", e.key)
}

// MemoryCache represents an in-memory cache implementation.
//
// MemoryCache supports cache item expiration. When adding a cache item, you can specify an expiration time or
// a validity period. If no expiration time is specified, the item is considered permanently valid.
// There is no limit to the number of cache items in MemoryCache. To avoid data race conditions in concurrent scenarios,
// MemoryCache employs locking mechanisms. As a result, having too many items can impact concurrent access performance.
// Additionally, to prevent frequent garbage collection scans and to avoid external modifications affecting the cache,
// MemoryCache stores the original content of the items, meaning that both storing and retrieving items
// involve copying values. It is also recommended that users do not cache pointers or excessively large items.
// Therefore, MemoryCache is not suitable for scenarios with a large number of cache items and
// high concurrent performance requirements.
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]Item
}

// NewMemoryCache creates and initializes a new MemoryCache instance.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]Item),
	}
}

// Get retrieves the value associated with the given key from the cache.
// If the key is not found or has expired, it returns an error.
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
	defer delete(c.items, keyGetter.GetKey())
	return nil, ErrKeyExpired{key: keyGetter.GetKey()}
}

// Set sets the value associated with the given key in the cache.
// It accepts optional ItemOption functions to customize the cache item.
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
	c.items[keyGetter.GetKey()] = *item
	return nil
}

// Delete removes the value associated with the given key from the cache.
func (c *MemoryCache) Delete(key KeyGetter) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key.GetKey())
	return nil
}

// Clear removes all items from the cache.
func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]Item)
	return nil
}
