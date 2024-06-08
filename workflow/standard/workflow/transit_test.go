// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package workflow

import (
	"context"
	"fmt"
	"testing"

	"github.com/rhosocial/go-dag/workflow/standard/cache"
	"github.com/stretchr/testify/assert"
)

// Mock implementations for testing
type mockCache struct{}

func (m *mockCache) Get(key cache.KeyGetter) (any, error) {
	return nil, nil
}

func (m *mockCache) Set(key cache.KeyGetter, value any, options ...cache.ItemOption) error {
	return nil
}

func (m *mockCache) Delete(key cache.KeyGetter) error {
	return nil
}

func (m *mockCache) Clear() error {
	return nil
}

type mockKeyGetter struct {
	keys []any
}

func (m *mockKeyGetter) GetKey() string {
	return fmt.Sprintf("%v", m.keys)
}

func (m *mockKeyGetter) GetKeyGetterFunc() KeyGetterFunc {
	return func(args ...any) cache.KeyGetter {
		return &mockKeyGetter{keys: args}
	}
}

// TestSimpleTransit_GetCache tests the GetCache method of SimpleTransit.
// It ensures that the correct cache instance is returned.
func TestSimpleTransit_GetCache(t *testing.T) {
	mockCacheInstance := &mockCache{}
	transit := SimpleTransit{
		cache: mockCacheInstance,
	}
	assert.Equal(t, mockCacheInstance, transit.GetCache())
}

// TestSimpleTransit_GetCacheEnabled tests the GetCacheEnabled method of SimpleTransit.
// It ensures that the method returns true if caching is enabled, and false otherwise.
func TestSimpleTransit_GetCacheEnabled(t *testing.T) {
	transit := SimpleTransit{}
	assert.False(t, transit.GetCacheEnabled())

	transit.cache = &mockCache{}
	assert.True(t, transit.GetCacheEnabled())
}

// TestSimpleTransit_GetKeyGetterFunc tests the GetKeyGetterFunc method of SimpleTransit.
// It ensures that the method returns nil if no KeyGetter is set, and returns a valid function otherwise.
func TestSimpleTransit_GetKeyGetterFunc(t *testing.T) {
	transit := SimpleTransit{}
	assert.Nil(t, transit.GetKeyGetterFunc())

	mockKeyGetterInstance := &mockKeyGetter{}
	transit.KeyGetter = mockKeyGetterInstance
	assert.NotNil(t, transit.GetKeyGetterFunc())
}

// TestNewTransit tests the NewTransit function.
// It ensures that a SimpleTransit is correctly created with the given parameters.
func TestNewTransit(t *testing.T) {
	name := "testTransit"
	incoming := []string{"in"}
	outgoing := []string{"out"}
	worker := func(ctx context.Context, a ...any) (any, error) { return a[0], nil }

	transit := NewTransit(name, incoming, outgoing, worker).(*SimpleTransit)

	assert.Equal(t, name, transit.GetName())
	assert.Equal(t, incoming, transit.GetIncoming())
	assert.Equal(t, outgoing, transit.GetOutgoing())
	assert.Nil(t, transit.cache)
	assert.Nil(t, transit.KeyGetter)
}

// TestTransitWithCacheKeyGetter tests the TransitWithCacheKeyGetter option.
// It ensures that the KeyGetter is correctly set in SimpleTransit.
func TestTransitWithCacheKeyGetter(t *testing.T) {
	transit := SimpleTransit{}
	mockKeyGetterInstance := &mockKeyGetter{}
	option := TransitWithCacheKeyGetter(mockKeyGetterInstance)
	option(&transit)

	assert.Equal(t, mockKeyGetterInstance, transit.KeyGetter)
}

// TestTransitWithCache tests the TransitWithCache option.
// It ensures that the cache is correctly set in SimpleTransit.
func TestTransitWithCache(t *testing.T) {
	transit := SimpleTransit{}
	mockCacheInstance := &mockCache{}
	option := TransitWithCache(mockCacheInstance)
	option(&transit)

	assert.Equal(t, mockCacheInstance, transit.cache)
}

// TestNewTransitWithOptions tests the NewTransit function with options.
// It ensures that the options (cache and KeyGetter) are correctly applied to SimpleTransit.
func TestNewTransitWithOptions(t *testing.T) {
	name := "testTransit"
	incoming := []string{"in"}
	outgoing := []string{"out"}
	worker := func(ctx context.Context, a ...any) (any, error) { return a[0], nil }

	mockCacheInstance := &mockCache{}
	mockKeyGetterInstance := &mockKeyGetter{}

	transit := NewTransit(name, incoming, outgoing, worker,
		TransitWithCache(mockCacheInstance),
		TransitWithCacheKeyGetter(mockKeyGetterInstance)).(*SimpleTransit)

	assert.Equal(t, mockCacheInstance, transit.GetCache())
	assert.True(t, transit.GetCacheEnabled())
	assert.Equal(t, mockKeyGetterInstance, transit.KeyGetter)
	assert.NotNil(t, transit.GetKeyGetterFunc())
}
