// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

// Package cache defines the cache interface and the simplest cache implementation.
package cache

// KeyGetter defines a method to retrieve keys.
type KeyGetter interface {
	GetKey() string
}

// Interface defines a generic caching interface for getting, setting, deleting, and clearing caches.
type Interface interface {
	// Get retrieves the value of the specified key from the cache.
	//
	// Parameter
	//
	// - key: an object implementing the KeyGetter interface to retrieve the key to get.
	//
	// Returns
	//
	// - any: the retrieved value, returned as an empty interface.
	//
	// - error: if an error occurs, it returns the corresponding error message.
	Get(key KeyGetter) (any, error)

	// Set sets the value of the specified key in the cache.
	//
	// Parameter
	//
	// - key: an object implementing the KeyGetter interface to retrieve the key to set.
	//
	// - value: the value to set, passed as an empty interface.
	//
	// - options: optional options to set cache items.
	//
	// Returns
	//
	// - error: if an error occurs, it returns the corresponding error message.
	Set(key KeyGetter, value any, options ...ItemOption) error

	// Delete deletes the value of the specified key from the cache.
	//
	// Parameter
	//
	// - key: an object implementing the KeyGetter interface to retrieve the key to delete.
	//
	// Returns
	//
	// - error: if an error occurs, it returns the corresponding error message.
	Delete(key KeyGetter) error

	// Clear clears the cache, deleting all cache items.
	//
	// Returns
	//
	// - error: if an error occurs, it returns the corresponding error message.
	Clear() error
}
