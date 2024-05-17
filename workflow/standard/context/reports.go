// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package context

import (
	"sync"

	"github.com/rhosocial/go-dag/workflow/standard/cache"
)

// ReportsItem defines a function type that takes an ReportsInterface and returns an error.
// It is used for setting options on a Reports instance.
type ReportsItem func(o *Reports) error

// ReportsInterface defines an interface for getting and setting global and transit cache items.
type ReportsInterface interface {
	AddGlobal(key string, value any) error
	GetGlobal(key string) (value any, err error)
	AddTransit(transit string, key string, value any) error
	GetTransit(transit string, key string) (value any, err error)
}

// ReportKey represents a key used in the cache, implementing the KeyGetter interface.
type ReportKey struct {
	key string
	cache.KeyGetter
}

// GetKey returns the key as a string.
func (r ReportKey) GetKey() string {
	return r.key
}

// Reports implements the ReportsInterface, holding global and transit caches.
type Reports struct {
	global     cache.MemoryCache             // Global cache for storing key-value pairs
	transits   map[string]*cache.MemoryCache // Map for storing transit-specific caches
	muTransits sync.RWMutex                  // Mutex for synchronizing access to transits map

	ReportsInterface
}

// AddGlobal adds a key-value pair to the global cache.
func (r *Reports) AddGlobal(key string, value any) error {
	return r.global.Set(ReportKey{key: key}, value)
}

// GetGlobal retrieves a value from the global cache using the given key.
func (r *Reports) GetGlobal(key string) (any, error) {
	return r.global.Get(ReportKey{key: key})
}

// AddTransit adds a key-value pair to a transit-specific cache.
func (r *Reports) AddTransit(transit string, key string, value any) error {
	r.muTransits.Lock()
	defer r.muTransits.Unlock()

	if r.transits == nil {
		r.transits = make(map[string]*cache.MemoryCache)
	}

	if _, ok := r.transits[transit]; !ok {
		r.transits[transit] = cache.NewMemoryCache()
	}
	return r.transits[transit].Set(OptionKey{key: key}, value)
}

// GetTransit retrieves a value from a transit-specific cache using the given transit and key.
func (r *Reports) GetTransit(transit string, key string) (any, error) {
	r.muTransits.Lock()
	defer r.muTransits.Unlock()

	if r.transits == nil {
		r.transits = make(map[string]*cache.MemoryCache)
	}
	if _, ok := r.transits[transit]; !ok {
		r.transits[transit] = cache.NewMemoryCache()
	}
	return r.transits[transit].Get(OptionKey{key: key})
}

// NewReports creates a new Reports instance and applies the given ReportsItem functions to it.
// It returns an ReportsInterface or an error if any ReportsItem function fails.
func NewReports(options ...ReportsItem) (ReportsInterface, error) {
	result := &Reports{
		global: *cache.NewMemoryCache(),
	}
	for _, option := range options {
		if err := option(result); err != nil {
			return nil, err
		}
	}
	return result, nil
}
