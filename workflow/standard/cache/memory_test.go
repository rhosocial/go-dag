// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package cache

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MemoryCacheKeyGetter struct {
	key1 uint
	key2 uint
	KeyGetter
}

func (k MemoryCacheKeyGetter) GetKey() string {
	return fmt.Sprintf("%d", k.key1<<32+k.key2)
}

func TestMemoryCache_SetAndGet(t *testing.T) {
	c := NewMemoryCache()
	if c == nil {
		t.Fail()
		return
	}
	assert.Empty(t, c.items)

	result, err := c.Get(MemoryCacheKeyGetter{key1: 1, key2: 2})
	assert.ErrorAs(t, err, &ErrKeyNotFound{})
	assert.Nil(t, result)

	err = c.Set(MemoryCacheKeyGetter{key1: 1, key2: 2}, int64(1))
	if err != nil {
		t.Fail()
		return
	}

	result, err = c.Get(MemoryCacheKeyGetter{key1: 1, key2: 2})
	if err != nil {
		t.Fail()
		return
	}
	assert.NotNil(t, result)
	i, ok := result.(int64)
	assert.True(t, ok)
	assert.Equal(t, int64(1), i)
}

func TestMemoryCache_Expiration(t *testing.T) {
	c := NewMemoryCache()
	assert.Len(t, c.items, 0)

	err := c.Set(MemoryCacheKeyGetter{key1: 1, key2: 2}, int64(1), WithExpiredAt(time.Now().Add(-time.Millisecond)))
	assert.ErrorIs(t, err, ErrInvalidExpiration{})
	assert.Equal(t, err.Error(), ErrInvalidExpiration{}.Error())
	assert.Len(t, c.items, 0)

	err = c.Set(MemoryCacheKeyGetter{key1: 1, key2: 2}, int64(1), WithExpiredAt(time.Now().Add(time.Millisecond)))
	assert.NoError(t, err)
	assert.Len(t, c.items, 1)

	time.Sleep(time.Millisecond)
	assert.Len(t, c.items, 1)

	result, err := c.Get(MemoryCacheKeyGetter{key1: 1, key2: 2})
	assert.ErrorAs(t, err, &ErrKeyExpired{key: MemoryCacheKeyGetter{key1: 1, key2: 2}.GetKey()})
	assert.Equal(t, err.Error(), ErrKeyExpired{key: MemoryCacheKeyGetter{key1: 1, key2: 2}.GetKey()}.Error())
	assert.Nil(t, result)
	assert.Len(t, c.items, 0)

	err = c.Set(MemoryCacheKeyGetter{key1: 1, key2: 2}, int64(1), WithTTL(-time.Millisecond))

	assert.Len(t, c.items, 0)
	assert.ErrorIs(t, err, ErrInvalidTTL{})
	assert.Equal(t, err.Error(), ErrInvalidTTL{}.Error())

	err = c.Set(MemoryCacheKeyGetter{key1: 1, key2: 2}, int64(1), WithTTL(time.Millisecond))
	assert.NoError(t, err)

	time.Sleep(time.Millisecond)
	assert.Len(t, c.items, 1)

	result, err = c.Get(MemoryCacheKeyGetter{key1: 1, key2: 2})
	assert.ErrorAs(t, err, &ErrKeyExpired{})
	assert.Nil(t, result)
	assert.Len(t, c.items, 0)
}

func TestMemoryCache_Delete(t *testing.T) {
	c := NewMemoryCache()

	err := c.Set(MemoryCacheKeyGetter{key1: 1, key2: 2}, int64(1), WithTTL(time.Millisecond))
	assert.NoError(t, err)

	err = c.Delete(MemoryCacheKeyGetter{key1: 1, key2: 3})
	if err != nil {
		t.Fail()
		return
	}

	assert.Len(t, c.items, 1)
	time.Sleep(time.Millisecond)

	result, err := c.Get(MemoryCacheKeyGetter{key1: 1, key2: 2})
	assert.ErrorIs(t, err, ErrKeyExpired{key: MemoryCacheKeyGetter{key1: 1, key2: 2}.GetKey()})
	assert.Nil(t, result)
	assert.Len(t, c.items, 0)
}

func TestMemoryCache_Clear(t *testing.T) {
	c := NewMemoryCache()

	err := c.Set(MemoryCacheKeyGetter{key1: 1, key2: 2}, int64(1), WithTTL(time.Millisecond))
	assert.NoError(t, err)
	assert.Len(t, c.items, 1)

	err = c.Clear()
	assert.Len(t, c.items, 0)

	result, err := c.Get(MemoryCacheKeyGetter{key1: 1, key2: 2})
	assert.ErrorIs(t, err, ErrKeyNotFound{key: MemoryCacheKeyGetter{key1: 1, key2: 2}.GetKey()})
	assert.Equal(t, err.Error(), ErrKeyNotFound{key: MemoryCacheKeyGetter{key1: 1, key2: 2}.GetKey()}.Error())
	assert.Nil(t, result)
	assert.Len(t, c.items, 0)
}

type testKey string

func (k testKey) GetKey() string {
	return string(k)
}

func benchmarkCacheSet(c Interface, b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := testKey("key" + strconv.Itoa(i))
		value := "value" + strconv.Itoa(i)
		c.Set(key, value)
	}
}

func benchmarkCacheGet(c Interface, b *testing.B) {
	// Pre-fill the cache
	for i := 0; i < b.N; i++ {
		key := testKey("key" + strconv.Itoa(i))
		value := "value" + strconv.Itoa(i)
		c.Set(key, value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := testKey("key" + strconv.Itoa(i))
		c.Get(key)
	}
}

func benchmarkCacheDelete(c Interface, b *testing.B) {
	// Pre-fill the cache
	for i := 0; i < b.N; i++ {
		key := testKey("key" + strconv.Itoa(i))
		value := "value" + strconv.Itoa(i)
		c.Set(key, value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := testKey("key" + strconv.Itoa(i))
		c.Delete(key)
	}
}

func benchmarkCacheClear(c Interface, b *testing.B) {
	// Pre-fill the cache
	for i := 0; i < 1000; i++ {
		key := testKey("key" + strconv.Itoa(i))
		value := "value" + strconv.Itoa(i)
		c.Set(key, value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Clear()
	}
}

func BenchmarkMemoryCache_Set(b *testing.B) {
	cache := NewMemoryCache()
	benchmarkCacheSet(cache, b)
}

func BenchmarkMemoryCache_Get(b *testing.B) {
	cache := NewMemoryCache()
	benchmarkCacheGet(cache, b)
}

func BenchmarkMemoryCache_Delete(b *testing.B) {
	cache := NewMemoryCache()
	benchmarkCacheDelete(cache, b)
}

func BenchmarkMemoryCache_Clear(b *testing.B) {
	cache := NewMemoryCache()
	benchmarkCacheClear(cache, b)
}

func benchmarkCacheParallel(c Interface, b *testing.B) {
	// Step 1: Pre-fill the cache with some values
	for i := 0; i < 1000; i++ {
		key := testKey("key" + strconv.Itoa(i))
		value := "value" + strconv.Itoa(i)
		c.Set(key, value)
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Step 2: Access some hit-and-miss keys
			for j := 0; j < 500; j++ {
				key := testKey("key" + strconv.Itoa(j))
				c.Get(key)
			}
			for j := 1000; j < 1500; j++ {
				key := testKey("key" + strconv.Itoa(j))
				c.Get(key)
			}

			// Step 3: Delete some keys
			for j := 0; j < 500; j++ {
				key := testKey("key" + strconv.Itoa(j))
				c.Delete(key)
			}

			// Step 4: Add some new values
			for j := 1500; j < 2000; j++ {
				key := testKey("key" + strconv.Itoa(j))
				value := "value" + strconv.Itoa(j)
				c.Set(key, value)
			}

			// Step 5: Access some hit-and-miss keys again
			for j := 500; j < 1000; j++ {
				key := testKey("key" + strconv.Itoa(j))
				c.Get(key)
			}
			for j := 2000; j < 2500; j++ {
				key := testKey("key" + strconv.Itoa(j))
				c.Get(key)
			}
		}
	})
}

func BenchmarkMemoryCache_Parallel(b *testing.B) {
	cache := NewMemoryCache()
	benchmarkCacheParallel(cache, b)
}
