package cache

import (
	"fmt"
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
