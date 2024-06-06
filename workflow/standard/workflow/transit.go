// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package workflow

import (
	"github.com/rhosocial/go-dag/workflow/standard/cache"
	"github.com/rhosocial/go-dag/workflow/standard/channel"
)

// CacheInterface defines the methods required for an object to support caching.
type CacheInterface interface {
	// GetCacheEnabled returns true if caching is enabled, false otherwise.
	GetCacheEnabled() bool
	// GetCache returns the cache instance.
	GetCache() cache.Interface
}

// KeyGetterFunc is a function type that takes any number of arguments and returns a cache.KeyGetter.
type KeyGetterFunc func(...any) cache.KeyGetter

// KeyGetter defines the method required for obtaining a KeyGetter function.
type KeyGetter interface {
	// GetKeyGetterFunc returns a function that generates a cache key.
	GetKeyGetterFunc() KeyGetterFunc
}

// SimpleTransit is a structure that embeds channel.Transit, KeyGetter, and CacheInterface.
// It represents a transit with optional caching and key generation capabilities.
type SimpleTransit struct {
	channel.Transit                 // Embedding channel.Transit for transit functionalities.
	KeyGetter                       // Embedding KeyGetter for key generation functionalities.
	CacheInterface                  // Embedding CacheInterface for cache functionalities.
	cache           cache.Interface // The cache instance.
}

// GetCache returns the cache instance associated with the SimpleTransit.
func (s *SimpleTransit) GetCache() cache.Interface {
	return s.cache
}

// GetCacheEnabled returns true if a cache instance is associated with the SimpleTransit.
func (s *SimpleTransit) GetCacheEnabled() bool {
	return s.cache != nil
}

// GetKeyGetterFunc returns the KeyGetter function from the embedded KeyGetter interface.
func (s *SimpleTransit) GetKeyGetterFunc() KeyGetterFunc {
	if s.KeyGetter == nil {
		return nil
	}
	return s.KeyGetter.GetKeyGetterFunc()
}

// TransitOption defines a function type for configuring a SimpleTransit.
type TransitOption func(*SimpleTransit)

// NewTransit creates a new SimpleTransit with the given name, incoming and outgoing channels,
// and worker function. It also applies any provided TransitOptions.
func NewTransit(name string, incoming, outgoing []string, worker channel.WorkerFunc, options ...TransitOption) channel.Transit {
	transit := SimpleTransit{
		Transit: channel.NewTransit(name, incoming, outgoing, worker), // Initialize the embedded Transit.
	}
	// Apply each provided option to the transit.
	for _, option := range options {
		option(&transit)
	}
	return &transit
}

// TransitWithCacheKeyGetter returns a TransitOption that sets the KeyGetter for the SimpleTransit.
func TransitWithCacheKeyGetter(keyGetter KeyGetter) TransitOption {
	return func(transit *SimpleTransit) {
		transit.KeyGetter = keyGetter
	}
}

// TransitWithCache returns a TransitOption that sets the cache for the SimpleTransit.
func TransitWithCache(cache cache.Interface) TransitOption {
	return func(transit *SimpleTransit) {
		transit.cache = cache
	}
}
