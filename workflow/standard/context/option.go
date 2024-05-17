// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package context

import (
	"sync"

	"github.com/rhosocial/go-dag/workflow/standard/cache"
)

// OptionsItem defines a function type that takes an OptionsInterface and returns an error.
// It is used for setting options on an Options instance.
type OptionsItem func(o *Options) error

// OptionsInterface defines an interface for getting and setting global and transit cache items.
type OptionsInterface interface {
	GetGlobal(key string) (any, error)
	GetTransit(transit string, key string) (any, error)
	setGlobal(key string, value any) error
	setTransit(transit string, key string, value any) error
}

// OptionKey represents a key used in the cache, implementing the KeyGetter interface.
type OptionKey struct {
	key string
	cache.KeyGetter
}

// GetKey returns the key as a string.
func (k OptionKey) GetKey() string {
	return k.key
}

// Options is a struct that holds global and transit caches and provides methods to manipulate them.
type Options struct {
	global     cache.MemoryCache
	transits   map[string]*cache.MemoryCache
	muTransits sync.RWMutex
	OptionsInterface
}

// setGlobal sets a value in the global cache for the given key.
func (o *Options) setGlobal(key string, value any) error {
	return o.global.Set(OptionKey{key: key}, value)
}

// setTransit sets a value in the transit cache for the given transit and key.
// It ensures thread safety with a write lock.
func (o *Options) setTransit(transit string, key string, value any) error {
	o.muTransits.Lock()
	defer o.muTransits.Unlock()

	if o.transits == nil {
		o.transits = make(map[string]*cache.MemoryCache)
	}

	if _, ok := o.transits[transit]; !ok {
		o.transits[transit] = cache.NewMemoryCache()
	}
	return o.transits[transit].Set(OptionKey{key: key}, value)
}

// GetGlobal retrieves a value from the global cache for the given key.
func (o *Options) GetGlobal(key string) (any, error) {
	return o.global.Get(OptionKey{key: key})
}

// GetTransit retrieves a value from the transit cache for the given transit and key.
// It ensures thread safety with a write lock and initializes the cache if necessary.
func (o *Options) GetTransit(transit string, key string) (any, error) {
	o.muTransits.Lock()
	defer o.muTransits.Unlock()

	if o.transits == nil {
		o.transits = make(map[string]*cache.MemoryCache)
	}
	if _, ok := o.transits[transit]; !ok {
		o.transits[transit] = cache.NewMemoryCache()
	}
	return o.transits[transit].Get(OptionKey{key: key})
}

// NewOptions creates a new Options instance and applies the given OptionsItem functions to it.
// It returns an OptionsInterface or an error if any OptionsItem function fails.
func NewOptions(options ...OptionsItem) (OptionsInterface, error) {
	result := &Options{
		global: *cache.NewMemoryCache(),
	}
	for _, option := range options {
		if err := option(result); err != nil {
			return nil, err
		}
	}
	return result, nil
}
